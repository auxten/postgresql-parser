# Copyright 2014 The Cockroach Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
# implied. See the License for the specific language governing
# permissions and limitations under the License.

# WARNING: This Makefile is not easily understood. If you're here looking for
# typical Make invocations to build the project and run tests, you'll be better
# served by running `make help`.
#
# Maintainers: the output of `make help` is automatically generated from the
# double-hash (##) comments throughout this Makefile. Please submit
# improvements!

# We need to define $(GO) early because it's needed for defs.mk.
GO      ?= go
# xgo is needed also for defs.mk.
override xgo := GOFLAGS= $(GO)

## Nearly everything below this point needs to have the vendor directory ready
## for use and will depend on bin/.submodules-initialized. In order to
## ensure this is available before the first "include" directive depending
## on it, we'll have it listed first thing.
##
## Note how the actions for this rule are *not* using $(GIT_DIR) which
## is otherwise defined in defs.mk above. This is because submodules
## are used in the process of definining the .mk files included by the
## Makefile, so it is not yet defined by the time
## `.submodules-initialized` is needed during a fresh build after a
## checkout.
.SECONDARY: bin/.submodules-initialized
bin/.submodules-initialized:
	gitdir=$$(git rev-parse --git-dir 2>/dev/null || true); \
	if test -n "$$gitdir"; then \
	   git submodule update --init --recursive; \
	fi
	mkdir -p $(@D)
	touch $@

# If the user wants to persist customizations for some variables, they
# can do so by defining `customenv.mk` in their work tree.
-include customenv.mk

ifeq "$(findstring bench,$(MAKECMDGOALS))" "bench"
$(if $(TESTS),$(error TESTS cannot be specified with `make bench` (did you mean BENCHES?)))
else
$(if $(BENCHES),$(error BENCHES can only be specified with `make bench`))
endif

# Prevent invoking make with a specific test name without a constraining
# package.
ifneq "$(filter bench% test% stress%,$(MAKECMDGOALS))" ""
ifeq "$(PKG)" ""
$(if $(subst -,,$(TESTS)),$(error TESTS must be specified with PKG (e.g. PKG=./pkg/sql)))
$(if $(subst -,,$(BENCHES)),$(error BENCHES must be specified with PKG (e.g. PKG=./pkg/sql)))
endif
endif

TYPE :=
ifneq "$(TYPE)" ""
$(error Make no longer understands TYPE. Use 'build/builder.sh mkrelease $(subst release-,,$(TYPE))' instead)
endif

# dep-build is set to non-empty if the .d files should be included.
# This definition makes it empty when only the targets "help" and/or "clean"
# are specified.
build-with-dep-files := $(or $(if $(MAKECMDGOALS),,implicit-all),$(filter-out help clean,$(MAKECMDGOALS)))

## Which package to run tests against, e.g. "./pkg/foo".
PKG := ./pkg/...

## Tests to run for use with `make test` or `make check-libroach`.
TESTS := .

## Benchmarks to run for use with `make bench`.
BENCHES :=

## Space delimited list of logic test files to run, for make testlogic/testccllogic/testoptlogic.
FILES :=

## Name of a logic test configuration to run, for make testlogic/testccllogic/testoptlogic.
## (default: all configs. It's not possible yet to specify multiple configs in this way.)
TESTCONFIG :=

## Regex for matching logic test subtests. This is always matched after "FILES"
## if they are provided.
SUBTESTS :=

## Test timeout to use for the linter.
LINTTIMEOUT := 20m

## Test timeout to use for regular tests.
TESTTIMEOUT := 30m

## Test timeout to use for race tests.
RACETIMEOUT := 30m

## Test timeout to use for acceptance tests.
ACCEPTANCETIMEOUT := 30m

## Test timeout to use for benchmarks.
BENCHTIMEOUT := 5m

## Extra flags to pass to the go test runner, e.g. "-v --vmodule=raft=1"
TESTFLAGS :=

## Flags to pass to `go test` invocations that actually run tests, but not
## elsewhere. Used for the -json flag which we'll only want to pass
## selectively.  There's likely a better solution.
GOTESTFLAGS :=

## Extra flags to pass to `stress` during `make stress`.
STRESSFLAGS :=

## Cluster to use for `make roachprod-stress`
CLUSTER :=

## Verbose allows turning on verbose output from the cmake builds.
VERBOSE :=

## Indicate the base root directory where to install
DESTDIR :=

DUPLFLAGS    := -t 100
GOFLAGS      :=
TAGS         :=
ARCHIVE      := cockroach.src.tgz
STARTFLAGS   := -s type=mem,size=1GiB --logtostderr
BUILDTARGET  := ./pkg/cmd/cockroach
SUFFIX       := $(GOEXE)
INSTALL      := install
prefix       := /usr/local
bindir       := $(prefix)/bin

ifeq "$(findstring -j,$(shell ps -o args= $$PPID))" ""
ifdef NCPUS
MAKEFLAGS += -j$(NCPUS)
$(info Running make with -j$(NCPUS))
endif
endif

BUILDTYPE := development

# Build C/C++ with basic debugging information.
CFLAGS += -g1
CXXFLAGS += -g1
LDFLAGS ?=

# TODO(benesch): remove filter-outs below when golang/go#26144 and
# golang/go#16651, respectively, are fixed.
CGO_CFLAGS = $(filter-out -g%,$(CFLAGS))
CGO_CXXFLAGS = $(CXXFLAGS)
CGO_LDFLAGS = $(filter-out -static,$(LDFLAGS))

export CFLAGS CXXFLAGS LDFLAGS CGO_CFLAGS CGO_CXXFLAGS CGO_LDFLAGS

# We intentionally use LINKFLAGS instead of the more traditional LDFLAGS
# because LDFLAGS has built-in semantics that don't make sense with the Go
# toolchain.
override LINKFLAGS = -X github.com/cockroachdb/cockroach/pkg/build.typ=$(BUILDTYPE) -extldflags "$(LDFLAGS)"

GOFLAGS ?=
TAR     ?= tar

# Ensure we have an unambiguous GOPATH.
GOPATH := $(shell $(GO) env GOPATH)

# We install our vendored tools to a directory within this repository to avoid
# overwriting any user-installed binaries of the same name in the default GOBIN.
GO_INSTALL := GOBIN='$(abspath bin)' GOFLAGS= $(GO) install

# Prefer tools we've installed with go install and Yarn to those elsewhere on
# the PATH.
export PATH := $(abspath bin):$(PATH)

# HACK: Make has a fast path and a slow path for command execution,
# but the fast path uses the PATH variable from when make was started,
# not the one we set on the previous line. In order for the above
# line to have any effect, we must force make to always take the slow path.
# Setting the SHELL variable to a value other than the default (/bin/sh)
# is one way to do this globally.
# http://stackoverflow.com/questions/8941110/how-i-could-add-dir-to-path-in-makefile/13468229#13468229
#
# We also force the PWD environment variable to $(CURDIR), which ensures that
# any programs invoked by Make see a physical CWD without any symlinks. The Go
# toolchain does not support symlinks well (for one example, see
# https://github.com/golang/go/issues/24359). This may be fixed when GOPATH is
# deprecated, so revisit whether this workaround is necessary then.
export SHELL := env PWD=$(CURDIR) bash
ifeq ($(SHELL),)
$(error bash is required)
endif

# Invocation of any NodeJS script should be prefixed by NODE_RUN. See the
# comments within node-run.sh for rationale.
NODE_RUN := build/node-run.sh

# make-lazy converts a recursive variable, which is evaluated every time it's
# referenced, to a lazy variable, which is evaluated only the first time it's
# used. See: http://blog.jgc.org/2016/07/lazy-gnu-make-variables.html
override make-lazy = $(eval $1 = $$(eval $1 := $(value $1))$$($1))

# GNU tar and BSD tar both support transforming filenames according to a regular
# expression, but have different flags to do so.
TAR_XFORM_FLAG = $(shell $(TAR) --version | grep -q GNU && echo "--xform='flags=r;s'" || echo "-s")
$(call make-lazy,TAR_XFORM_FLAG)

# To edit in-place without creating a backup file, GNU sed requires a bare -i,
# while BSD sed requires an empty string as the following argument.
SED_INPLACE = sed $(shell sed --version 2>&1 | grep -q GNU && echo -i || echo "-i ''")
$(call make-lazy,SED_INPLACE)

# MAKE_TERMERR is set automatically in Make v4.1+, but macOS is still shipping
# v3.81.
MAKE_TERMERR ?= $(shell [[ -t 2 ]] && echo true)

# This is how you get a literal space into a Makefile.
space := $(eval) $(eval)

# Color support.
yellow = $(shell { tput setaf 3 || tput AF 3; } 2>/dev/null)
cyan = $(shell { tput setaf 6 || tput AF 6; } 2>/dev/null)
term-reset = $(shell { tput sgr0 || tput me; } 2>/dev/null)
$(call make-lazy,yellow)
$(call make-lazy,cyan)
$(call make-lazy,term-reset)

# Tell Make to delete the target if its recipe fails. Otherwise, if a recipe
# modifies its target before failing, the target's timestamp will make it appear
# up-to-date on the next invocation of Make, even though it is likely corrupt.
# See: https://www.gnu.org/software/make/manual/html_node/Errors.html#Errors
.DELETE_ON_ERROR:

# Targets that name a real file that must be rebuilt on every Make invocation
# should depend on .ALWAYS_REBUILD. (.PHONY should only be used on targets that
# don't name a real file because .DELETE_ON_ERROR does not apply to .PHONY
# targets.)
.ALWAYS_REBUILD:
.PHONY: .ALWAYS_REBUILD

## Update the git hooks and install commands from dependencies whenever they
## change.
bin/.bootstrap: bin/.submodules-initialized
	echo "skip"
	touch $@

IGNORE_GOVERS :=

# The following section handles building our C/C++ dependencies. These are
# common because both the root Makefile and protobuf.mk have C dependencies.

host-is-macos := $(findstring Darwin,$(UNAME))
host-is-mingw := $(findstring MINGW,$(UNAME))

ifdef host-is-macos
# On macOS 10.11, XCode SDK v8.1 (and possibly others) indicate the presence of
# symbols that don't exist until macOS 10.12. Setting MACOSX_DEPLOYMENT_TARGET
# to the host machine's actual macOS version works around this. See:
# https://github.com/jemalloc/jemalloc/issues/494.
export MACOSX_DEPLOYMENT_TARGET ?= $(macos-version)
endif

# Cross-compilation occurs when you set TARGET_TRIPLE to something other than
# HOST_TRIPLE. You'll need to ensure the cross-compiling toolchain is on your
# path and override the rest of the variables that immediately follow as
# necessary. For an example, see build/builder/cmd/mkrelease, which sets these
# variables appropriately for the toolchains baked into the builder image.
TARGET_TRIPLE := $(HOST_TRIPLE)
XCMAKE_SYSTEM_NAME :=
XGOOS :=
XGOARCH :=
XCC := $(TARGET_TRIPLE)-cc
XCXX := $(TARGET_TRIPLE)-c++
EXTRA_XCMAKE_FLAGS :=
EXTRA_XCONFIGURE_FLAGS :=

ifneq ($(HOST_TRIPLE),$(TARGET_TRIPLE))
is-cross-compile := 1
endif

target-is-windows := $(findstring w64,$(TARGET_TRIPLE))

# CMAKE_TARGET_MESSAGES=OFF prevents CMake from printing progress messages
# whenever a target is fully built to prevent spammy output from make when
# c-deps are all already built. Progress messages are still printed when actual
# compilation is being performed.
cmake-flags := -DCMAKE_TARGET_MESSAGES=OFF $(if $(host-is-mingw),-G 'MSYS Makefiles')
configure-flags :=

# Use xcmake-flags when invoking CMake on libraries/binaries for the target
# platform (i.e., the cross-compiled platform, if specified); use plain
# cmake-flags when invoking CMake on libraries/binaries for the host platform.
# Similarly for xconfigure-flags and configure-flags, and xgo and GO.
xcmake-flags := $(cmake-flags) $(EXTRA_XCMAKE_FLAGS)
xconfigure-flags := $(configure-flags) $(EXTRA_XCONFIGURE_FLAGS)

# If we're cross-compiling, inform Autotools and CMake.
ifdef is-cross-compile
xconfigure-flags += --host=$(TARGET_TRIPLE) CC=$(XCC) CXX=$(XCXX)
xcmake-flags += -DCMAKE_SYSTEM_NAME=$(XCMAKE_SYSTEM_NAME) -DCMAKE_C_COMPILER=$(XCC) -DCMAKE_CXX_COMPILER=$(XCXX)
override xgo := GOFLAGS= GOOS=$(XGOOS) GOARCH=$(XGOARCH) CC=$(XCC) CXX=$(XCXX) $(xgo)
endif

# Derived build variants.
use-stdmalloc          := $(findstring stdmalloc,$(TAGS))
use-msan               := $(findstring msan,$(GOFLAGS))

# User-requested build variants.
ENABLE_LIBROACH_ASSERTIONS ?=
ENABLE_ROCKSDB_ASSERTIONS ?=

BUILD_DIR := $(GOPATH)/native/$(TARGET_TRIPLE)

# In MinGW, cgo flags don't handle Unix-style paths, so convert our base path to
# a Windows-style path.
#
# TODO(benesch): Figure out why. MinGW transparently converts Unix-style paths
# everywhere else.
ifdef host-is-mingw
BUILD_DIR := $(shell cygpath -m $(BUILD_DIR))
endif

# In each package that uses cgo, we inject include and library search paths into
# files named zcgo_flags_{native-tag}.go. The logic for this is complicated so
# that Make-driven builds can cache the state of builds for multiple
# configurations at once, while still allowing the use of `go build` and `go
# test` for the configuration most recently built with Make.
#
# Building with Make always adds the `make` and {native-tag} tags to the build.
#
# Unsuffixed flags files (zcgo_flags.cgo) have the build constraint `!make` and
# are only compiled when invoking the Go toolchain directly on a package-- i.e.,
# when the `make` build tag is not specified. These files are rebuilt whenever
# the build signature changes (see build/defs.mk.sig), and so reflect the target
# triple that Make was most recently invoked with.
#
# Suffixed flags files (e.g. zcgo_flags_{native-tag}.go) have the build
# constraint `{native-tag}` and are built the first time a Make-driven build
# encounters a given native tag or when the build signature changes (see
# build/defs.mk.sig). These tags are unset when building with the Go toolchain
# directly, so these files are only compiled when building with Make.
CGO_PKGS := \
	pkg/cli \
	pkg/server/status \
	pkg/storage \
	pkg/ccl/storageccl/engineccl \
	pkg/ccl/gssapiccl \
	vendor/github.com/knz/go-libedit/unix
vendor/github.com/knz/go-libedit/unix-package := libedit_unix

$(BASE_CGO_FLAGS_FILES): Makefile build/defs.mk.sig | bin/.submodules-initialized
	@echo "regenerating $@"
	@echo '// GENERATED FILE DO NOT EDIT' > $@
	@echo >> $@
	@echo '// +build $(if $(findstring $(native-tag),$@),$(native-tag),!make)' >> $@
	@echo >> $@
	@echo 'package $(if $($(@D)-package),$($(@D)-package),$(notdir $(@D)))' >> $@
	@echo >> $@
	@echo '// #cgo CPPFLAGS: $(addprefix -I,$(JEMALLOC_DIR)/include $(KRB_CPPFLAGS))' >> $@
	@echo '// #cgo LDFLAGS: $(addprefix -L,$(CRYPTOPP_DIR) $(PROTOBUF_DIR) $(JEMALLOC_DIR)/lib $(SNAPPY_DIR) $(LIBEDIT_DIR)/src/.libs $(ROCKSDB_DIR) $(LIBROACH_DIR) $(KRB_DIR))' >> $@
	@echo 'import "C"' >> $@

# Convenient names for maintainers. Not used by other targets in the Makefile.
.PHONY: protoc libcryptopp libjemalloc libprotobuf libsnappy librocksdb libroach libroachccl libkrb5
protoc:      $(PROTOC)

PHONY: check-libroach
check-libroach: ## Run libroach tests.
check-libroach: $(LIBROACH_DIR)/Makefile $(LIBJEMALLOC) $(LIBPROTOBUF) $(LIBSNAPPY) $(LIBROCKSDB) $(LIBCRYPTOPP)
	@$(MAKE) --no-print-directory -C $(LIBROACH_DIR)
	cd $(LIBROACH_DIR) && ctest -V -R $(TESTS)

override TAGS += make $(native-tag)

# Some targets (protobuf) produce different results depending on the sort order;
# set LC_ALL so this is consistent across systems.
export LC_ALL=C

# defs.mk.sig attempts to capture common cases where defs.mk needs to be
# recomputed, like when compiling for a different platform or using a different
# Go binary. It is not intended to be perfect. Upgrading the compiler toolchain
# in place will go unnoticed, for example. Similar problems exist in all Make-
# based build systems and are not worth solving.
build/defs.mk.sig: sig = $(PATH):$(CURDIR):$(GO):$(GOPATH):$(CC):$(CXX):$(TARGET_TRIPLE):$(BUILDTYPE):$(IGNORE_GOVERS):$(ENABLE_LIBROACH_ASSERTIONS):$(ENABLE_ROCKSDB_ASSERTIONS)
build/defs.mk.sig: .ALWAYS_REBUILD
	@echo '$(sig)' | cmp -s - $@ || echo '$(sig)' > $@

COCKROACH      := ./cockroach$(SUFFIX)
COCKROACHOSS   := ./cockroachoss$(SUFFIX)
COCKROACHSHORT := ./cockroachshort$(SUFFIX)

SQLPARSER_TARGETS = \
	pkg/sql/parser/sql.go \
	pkg/sql/parser/helpmap_test.go \
	pkg/sql/parser/help_messages.go \
	pkg/sql/lex/tokens.go \
	pkg/sql/lex/keywords.go \
	pkg/sql/lex/reserved_keywords.go

#PROTOBUF_TARGETS := bin/.go_protobuf_sources bin/.gw_protobuf_sources bin/.cpp_protobuf_sources bin/.cpp_ccl_protobuf_sources

DOCGEN_TARGETS := bin/.docgen_bnfs bin/.docgen_functions

EXECGEN_TARGETS = \
  pkg/col/coldata/vec.eg.go \
  pkg/sql/colexec/and_or_projection.eg.go \
  pkg/sql/colexec/any_not_null_agg.eg.go \
  pkg/sql/colexec/avg_agg.eg.go \
  pkg/sql/colexec/bool_and_or_agg.eg.go \
  pkg/sql/colexec/cast.eg.go \
  pkg/sql/colexec/const.eg.go \
  pkg/sql/colexec/count_agg.eg.go \
  pkg/sql/colexec/distinct.eg.go \
  pkg/sql/colexec/hashjoiner.eg.go \
  pkg/sql/colexec/hashtable.eg.go \
  pkg/sql/colexec/hash_aggregator.eg.go \
  pkg/sql/colexec/hash_utils.eg.go \
  pkg/sql/colexec/like_ops.eg.go \
  pkg/sql/colexec/mergejoinbase.eg.go \
  pkg/sql/colexec/mergejoiner_fullouter.eg.go \
  pkg/sql/colexec/mergejoiner_inner.eg.go \
  pkg/sql/colexec/mergejoiner_leftanti.eg.go \
  pkg/sql/colexec/mergejoiner_leftouter.eg.go \
  pkg/sql/colexec/mergejoiner_leftsemi.eg.go \
  pkg/sql/colexec/mergejoiner_rightouter.eg.go \
  pkg/sql/colexec/min_max_agg.eg.go \
  pkg/sql/colexec/orderedsynchronizer.eg.go \
  pkg/sql/colexec/overloads_test_utils.eg.go \
  pkg/sql/colexec/proj_const_left_ops.eg.go \
  pkg/sql/colexec/proj_const_right_ops.eg.go \
  pkg/sql/colexec/proj_non_const_ops.eg.go \
  pkg/sql/colexec/quicksort.eg.go \
  pkg/sql/colexec/rank.eg.go \
  pkg/sql/colexec/relative_rank.eg.go \
  pkg/sql/colexec/row_number.eg.go \
  pkg/sql/colexec/rowstovec.eg.go \
  pkg/sql/colexec/selection_ops.eg.go \
  pkg/sql/colexec/select_in.eg.go \
  pkg/sql/colexec/sort.eg.go \
  pkg/sql/colexec/substring.eg.go \
  pkg/sql/colexec/sum_agg.eg.go \
  pkg/sql/colexec/values_differ.eg.go \
  pkg/sql/colexec/vec_comparators.eg.go \
  pkg/sql/colexec/window_peer_grouper.eg.go

.PHONY: remove_obsolete_execgen
remove_obsolete_execgen:
	@obsolete="$(filter-out $(EXECGEN_TARGETS), $(shell find pkg/col/coldata pkg/sql/colexec pkg/sql/exec -name '*.eg.go' 2>/dev/null))"; \
	for file in $${obsolete}; do \
	  echo "Removing obsolete file $${file}..."; \
	  rm -f $${file}; \
	done

OPTGEN_TARGETS = \
	pkg/sql/opt/memo/expr.og.go \
	pkg/sql/opt/operator.og.go \
	pkg/sql/opt/xform/explorer.og.go \
	pkg/sql/opt/norm/factory.og.go \
	pkg/sql/opt/rule_name.og.go \
	pkg/sql/opt/rule_name_string.go

go-targets-ccl := \
	$(COCKROACH) $(COCKROACHSHORT) \
	bin/workload \
	go-install \
	bench benchshort \
	check test testshort testslow testrace testraceslow testbuild \
	stress stressrace \
	roachprod-stress roachprod-stressrace \
	generate \
	lint lintshort

go-targets := $(go-targets-ccl) $(COCKROACHOSS)

.DEFAULT_GOAL := all
all: build

.PHONY: c-deps
c-deps: $(C_LIBS_CCL)

build-mode = build -o $@

go-install: build-mode = install

#$(COCKROACH) go-install generate: pkg/ui/distccl/bindata.go

$(COCKROACHOSS): BUILDTARGET = ./pkg/cmd/cockroach-oss
#$(COCKROACHOSS): $(C_LIBS_OSS) pkg/ui/distoss/bindata.go

$(COCKROACHSHORT): BUILDTARGET = ./pkg/cmd/cockroach-short

$(go-targets-ccl): $(C_LIBS_CCL)

BUILDINFO = .buildinfo/tag .buildinfo/rev
BUILD_TAGGED_RELEASE =

## Override for .buildinfo/tag
BUILDINFO_TAG :=

$(go-targets): $(BUILDINFO) $(CGO_FLAGS_FILES) $(PROTOBUF_TARGETS)
#$(go-targets): $(SQLPARSER_TARGETS) $(EXECGEN_TARGETS) $(OPTGEN_TARGETS)
$(go-targets): $(SQLPARSER_TARGETS) $(OPTGEN_TARGETS)
$(go-targets): override LINKFLAGS += \
	-X "github.com/cockroachdb/cockroach/pkg/build.tag=$(if $(BUILDINFO_TAG),$(BUILDINFO_TAG),$(shell cat .buildinfo/tag))" \
	-X "github.com/cockroachdb/cockroach/pkg/build.rev=$(shell cat .buildinfo/rev)" \
	-X "github.com/cockroachdb/cockroach/pkg/build.cgoTargetTriple=$(TARGET_TRIPLE)" \
	$(if $(BUILDCHANNEL),-X "github.com/cockroachdb/cockroach/pkg/build.channel=$(BUILDCHANNEL)") \
	$(if $(BUILD_TAGGED_RELEASE),-X "github.com/cockroachdb/cockroach/pkg/util/log.crashReportEnv=$(if $(BUILDINFO_TAG),$(BUILDINFO_TAG),$(shell cat .buildinfo/tag))")

# The build.utcTime format must remain in sync with TimeFormat in
# pkg/build/info.go. It is not installed in tests to avoid busting the cache on
# every rebuild.
$(COCKROACH) $(COCKROACHOSS) $(COCKROACHSHORT) go-install: override LINKFLAGS += \
	-X "github.com/cockroachdb/cockroach/pkg/build.utcTime=$(shell date -u '+%Y/%m/%d %H:%M:%S')"

SETTINGS_DOC_PAGE := docs/generated/settings/settings.html

# Note: We pass `-v` to `go build` and `go test -i` so that warnings
# from the linker aren't suppressed. The usage of `-v` also shows when
# dependencies are rebuilt which is useful when switching between
# normal and race test builds.
.PHONY: go-install
$(COCKROACH) $(COCKROACHOSS) $(COCKROACHSHORT) go-install:
	 $(xgo) $(build-mode) -v $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LINKFLAGS)' $(BUILDTARGET)

build: $(COCKROACH)

# Do not use plumbing commands, like git diff-index, in this target. Our build
# process modifies files quickly enough that plumbing commands report false
# positives on filesystems with only one second of resolution as a performance
# optimization. Porcelain commands, like git diff, exist to detect and remove
# these false positives.
#
# For details, see the "Possible timestamp problems with diff-files?" thread on
# the Git mailing list (http://marc.info/?l=git&m=131687596307197).
.buildinfo/tag: | .buildinfo
	@{ git describe --tags --dirty --match=v[0-9]* 2> /dev/null || git rev-parse --short HEAD; } | tr -d \\n > $@

.buildinfo/rev: | .buildinfo
	@git rev-parse HEAD > $@

ifneq ($(GIT_DIR),)
# If we're in a Git checkout, we update the buildinfo information on every build
# to keep it up-to-date.
.buildinfo/tag: .ALWAYS_REBUILD
.buildinfo/rev: .ALWAYS_REBUILD
endif

.SECONDARY: pkg/sql/parser/gen/sql.go.tmp
pkg/sql/parser/gen/sql.go.tmp: pkg/sql/parser/gen/sql-gen.y bin/.bootstrap
	set -euo pipefail; \
	  ret=$$(cd pkg/sql/parser/gen && goyacc -p sql -o sql.go.tmp sql-gen.y); \
	  if expr "$$ret" : ".*conflicts" >/dev/null; then \
	    echo "$$ret"; exit 1; \
	  fi

# The lex package needs to know about all tokens, because the encode
# functions and lexing predicates need to know about keywords, and
# keywords map to the token constants. Therefore, generate the
# constant tokens in the lex package primarily.
pkg/sql/lex/tokens.go: pkg/sql/parser/gen/sql.go.tmp
	(echo "// Code generated by make. DO NOT EDIT."; \
	 echo "// GENERATED FILE DO NOT EDIT"; \
	 echo; \
	 echo "package lex"; \
	 echo; \
	 grep '^const [A-Z][_A-Z0-9]* ' $^) > $@.tmp || rm $@.tmp
	mv -f $@.tmp $@

# The lex package is now the primary source for the token constant
# definitions. Modify the code generated by goyacc here to refer to
# the definitions in the lex package.
pkg/sql/parser/sql.go: pkg/sql/parser/gen/sql.go.tmp | bin/.bootstrap
	(echo "// Code generated by goyacc. DO NOT EDIT."; \
	 echo "// GENERATED FILE DO NOT EDIT"; \
	 cat $^ | \
	 sed -E 's/^const ([A-Z][_A-Z0-9]*) =.*$$/const \1 = lex.\1/g') > $@.tmp || rm $@.tmp
	mv -f $@.tmp $@
	goimports -w $@

# This modifies the grammar to:
# - improve the types used by the generated parser for non-terminals
# - expand the help rules.
#
# For types:
# Determine the types that will be migrated to union types by looking
# at the accessors of sqlSymUnion. The first step in this pipeline
# prints every return type of a sqlSymUnion accessor on a separate line.
# The next step regular expression escapes these types. The third
# (prepending) and the fourth (appending) steps build regular expressions
# for each of the types and store them in the file. (We make multiple
# regular expressions because we ran into a limit of characters for a
# single regex executed by sed.)
# Then translate the original syntax file, with the types determined
# above being replaced with the union type in their type declarations.
.SECONDARY: pkg/sql/parser/gen/sql-gen.y
pkg/sql/parser/gen/sql-gen.y: pkg/sql/parser/sql.y pkg/sql/parser/replace_help_rules.awk
	mkdir -p pkg/sql/parser/gen
	set -euo pipefail; \
	awk '/func.*sqlSymUnion/ {print $$(NF - 1)}' pkg/sql/parser/sql.y | \
	sed -e 's/[]\/$$*.^|[]/\\&/g' | \
	sed -e "s/^/s_(type|token) <(/" | \
	awk '{print $$0")>_\\1 <union> /* <\\2> */_"}' > pkg/sql/parser/gen/types_regex.tmp; \
	sed -E -f pkg/sql/parser/gen/types_regex.tmp < pkg/sql/parser/sql.y | \
	awk -f pkg/sql/parser/replace_help_rules.awk | \
	sed -Ee 's,//.*$$,,g;s,/[*]([^*]|[*][^/])*[*]/, ,g;s/ +$$//g' > $@.tmp || rm $@.tmp
	mv -f $@.tmp $@
	rm pkg/sql/parser/gen/types_regex.tmp

pkg/sql/lex/reserved_keywords.go: pkg/sql/parser/sql.y pkg/sql/parser/reserved_keywords.awk | bin/.bootstrap
	awk -f pkg/sql/parser/reserved_keywords.awk < $< > $@.tmp || rm $@.tmp
	mv -f $@.tmp $@
	gofmt -s -w $@

pkg/sql/lex/keywords.go: pkg/sql/parser/sql.y pkg/sql/lex/all_keywords.go | bin/.bootstrap
	go run -tags all_keywords pkg/sql/lex/all_keywords.go < $< > $@.tmp || rm $@.tmp
	mv -f $@.tmp $@
	gofmt -s -w $@

# This target will print unreserved_keywords which are not actually
# used in the grammar.
.PHONY: sqlparser-unused-unreserved-keywords
sqlparser-unused-unreserved-keywords: pkg/sql/parser/sql.y pkg/sql/parser/unreserved_keywords.awk
	@for kw in $$(awk -f pkg/sql/parser/unreserved_keywords.awk < $<); do \
	  if [ $$(grep -c $${kw} $<) -le 2 ]; then \
	    echo $${kw}; \
	  fi \
	done

pkg/sql/parser/helpmap_test.go: pkg/sql/parser/gen/sql-gen.y pkg/sql/parser/help_gen_test.sh | bin/.bootstrap
	@pkg/sql/parser/help_gen_test.sh < $< >$@.tmp || rm $@.tmp
	mv -f $@.tmp $@
	gofmt -s -w $@

pkg/sql/parser/help_messages.go: pkg/sql/parser/sql.y pkg/sql/parser/help.awk | bin/.bootstrap
	awk -f pkg/sql/parser/help.awk < $< > $@.tmp || rm $@.tmp
	mv -f $@.tmp $@
	gofmt -s -w $@

bins = \
  bin/allocsim \
  bin/benchmark \
  bin/cockroach-oss \
  bin/cockroach-short \
  bin/docgen \
  bin/execgen \
  bin/fuzz \
  bin/generate-binary \
  bin/terraformgen \
  bin/github-post \
  bin/github-pull-request-make \
  bin/gossipsim \
  bin/langgen \
  bin/protoc-gen-gogoroach \
  bin/publish-artifacts \
  bin/publish-provisional-artifacts \
  bin/optgen \
  bin/returncheck \
  bin/roachvet \
  bin/roachprod \
  bin/roachprod-stress \
  bin/roachtest \
  bin/teamcity-trigger \
  bin/uptodate \
  bin/urlcheck \
  bin/workload \
  bin/zerosum

testbins = \
  bin/logictest \
  bin/logictestopt \
  bin/logictestccl

# Mappings for binaries that don't live in pkg/cmd.
execgen-package = ./pkg/sql/colexec/execgen/cmd/execgen
langgen-package = ./pkg/sql/opt/optgen/cmd/langgen
optgen-package = ./pkg/sql/opt/optgen/cmd/optgen
logictest-package = ./pkg/sql/logictest
logictestccl-package = ./pkg/ccl/logictestccl
logictestopt-package = ./pkg/sql/opt/exec/execbuilder
terraformgen-package = ./pkg/cmd/roachprod/vm/aws/terraformgen
logictest-bins := bin/logictest bin/logictestopt bin/logictestccl

# Additional dependencies for binaries that depend on generated code.
#
# TODO(benesch): Derive this automatically. This is getting out of hand.
bin/workload bin/docgen bin/execgen bin/roachtest $(logictest-bins): $(SQLPARSER_TARGETS) $(PROTOBUF_TARGETS)
bin/workload bin/roachtest $(logictest-bins): $(EXECGEN_TARGETS)
bin/roachtest $(logictest-bins): $(C_LIBS_CCL) $(CGO_FLAGS_FILES) $(OPTGEN_TARGETS)

$(bins): bin/%: bin/%.d | bin/prereqs bin/.submodules-initialized
	@echo go install -v $*
	bin/prereqs $(if $($*-package),$($*-package),./pkg/cmd/$*) > $@.d.tmp
	mv -f $@.d.tmp $@.d
	@$(GO_INSTALL) -v $(if $($*-package),$($*-package),./pkg/cmd/$*)

$(testbins): bin/%: bin/%.d | bin/prereqs $(SUBMODULES_TARGET)
	@echo go test -c $($*-package)
	bin/prereqs -bin-name=$* -test $($*-package) > $@.d.tmp
	mv -f $@.d.tmp $@.d
	$(xgo) test $(GOTESTFLAGS) $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LINKFLAGS)' -c -o $@ $($*-package)

bin/prereqs: ./pkg/cmd/prereqs/*.go | bin/.submodules-initialized
	@echo go install -v ./pkg/cmd/prereqs
	@$(GO_INSTALL) -v ./pkg/cmd/prereqs

.PHONY: fuzz
fuzz: ## Run fuzz tests.
fuzz: bin/fuzz
	bin/fuzz $(TESTFLAGS) -tests $(TESTS) -timeout $(TESTTIMEOUT) $(PKG)


# No need to include all the dependency files if the user is just
# requesting help or cleanup.
ifneq ($(build-with-dep-files),)
.SECONDARY: bin/%.d
.PRECIOUS: bin/%.d
bin/%.d: ;

include $(wildcard bin/*.d)
endif
