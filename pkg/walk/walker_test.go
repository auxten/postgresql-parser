package walk

import (
	"sort"
	"testing"

	"github.com/auxten/postgresql-parser/pkg/sql/parser"
)

func TestParser(t *testing.T) {
	t.Run("SQL to AST", func(t *testing.T) {
		const sql = "SELECT SUBSTR(t1.TRANS_DATE, 0, 10) as trans_date, t1.TRANS_BRAN_CODE as trans_bran_code,ROUND(SUM(t1.TANS_AMT)/10000,2) as balance, count(t1.rowid) as cnt FROM mj t1 WHERE t1.MC_TRSCODE in ('INQ', 'LIS', 'CWD', 'CDP', 'TFR', 'PIN', 'REP', 'PAY') AND t1.TRANS_FLAG = '0' GROUP BY SUBSTR(t1.TRANS_DATE, 0, 10),t1.TRANS_BRAN_CODE ORDER by trans_date;"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}
		if ast == nil {
			t.Fatalf("got nil AST")
		}
		referredCols, err := ColNamesInSelect(sql)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		cols := referredCols.ToList()

		expectedCols := []string{"mc_trscode", "rowid", "tans_amt", "trans_bran_code", "trans_date", "trans_flag"}
		for i, v := range cols {
			if v != expectedCols[i] {
				t.Errorf("Expect %s, got %s", expectedCols, cols)
				break
			}
		}
	})
}

func allColsContained(set ReferredCols, cols []string) bool {
	if cols == nil {
		if set == nil {
			return true
		} else {
			return false
		}
	}
	if len(set) != len(cols) {
		return false
	}
	for _, col := range cols {
		if _, exist := set[col]; !exist {
			return false
		}
	}
	return true
}

func TestReferredVarsInSelectStatement(t *testing.T) {
	testCases := []struct {
		sql  string
		cols []string
		err  error
	}{
		{"SELECT a.r1, a.r2 FROM a ORDER BY a.r3 LIMIT 1", []string{"r1", "r2", "r3"}, nil},
		{"SELECT SUBSTR(t1.TRANS_DATE, 0, 10) as trans_date, t1.TRANS_BRAN_CODE as trans_bran_code,ROUND(SUM(t1.TANS_AMT)/10000,2) as balance, count(t1.rowid) as cnt FROM mj t1 WHERE t1.MC_TRSCODE in ('INQ', 'LIS', 'CWD', 'CDP', 'TFR', 'PIN', 'REP', 'PAY') AND t1.TRANS_FLAG = '0' GROUP BY SUBSTR(t1.TRANS_DATE, 0, 10),t1.TRANS_BRAN_CODE ORDER by trans_date;", []string{"mc_trscode", "rowid", "tans_amt", "trans_bran_code", "trans_date", "trans_flag"}, nil},
		{"SELECT count(DISTINCT s_i_id) FROM order_line JOIN stock ON s_i_id=ol_i_id AND s_w_id=ol_w_id WHERE ol_w_id = $1 AND ol_d_id = $2 AND ol_o_id BETWEEN $3 - 20 AND $3 - 1 AND s_quantity < $4", []string{"s_i_id", "ol_i_id", "s_w_id", "ol_w_id", "ol_d_id", "ol_o_id", "s_quantity"}, nil},
		{"SELECT c.i / j FROM a AS c JOIN b ON true;", []string{"i", "j"}, nil},
		{"WITH  with_t1 (c1, c2)   AS (    SELECT 1, 1/0   ) SELECT * FROM with_t1", []string{"*"}, nil},
		{`select marr
			from (select marr_stat_cd AS marr, label AS l
				  from root_loan_mock_v4
				  order by root_loan_mock_v4.age desc, l desc
				  limit 5) as v4
			LIMIT 1;`, []string{"age", "l", "label", "marr", "marr_stat_cd"}, nil},
		{"SELECT TAB_1222.COL1   AS  COL_3470 ,   TAB_1222.COL0   AS  COL_3471 ,   TAB_1222.COL3   AS  COL_3472 LIMIT 5", []string{"col1", "col0", "col3"}, nil},
		{`SELECT
				SET_MASKLEN('4ac:ded4:393a:a371:7690:9d0f:4817:3371/43', TAB_1226.COL5) AS  COL_3476
			FROM
				DEFAULTDB.PUBLIC.TABLE1 AS  TAB_1223
			RIGHT JOIN
				(
					SELECT
						TAB_1222.COL1   AS  COL_3470
					,   TAB_1222.COL0   AS  COL_3471
					,   TAB_1222.COL3   AS  COL_3472
					LIMIT 5
				)   AS  TAB_1224
			(COL_3473, COL_3474, COL_3475)
			JOIN
				DEFAULTDB.PUBLIC
			.TABLE0 AS  TAB_1225
			RIGHT JOIN
				DEFAULTDB.PUBLIC
			.TABLE1 AS  TAB_1226
			ON
				NULL
			FULL JOIN
				DEFAULTDB.PUBLIC
			.TABLE1 AS  TAB_1227
			ON
				FALSE   ON  SIMILAR_TO_ESCAPE(TAB_1222.COL6, NULL, TAB_1222.COL6)   ON
					INET_CONTAINS_OR_EQUALS(SET_MASKLEN('198f:60f5:287a:8163:c091:2a95:afdc:ae8b/108',
				(-4475677368810664623)), '6d38:61ce:1af7:9283:cf0d:beb2:23e0:d7f/109')
			ORDER BY
				TAB_1226.COL7   DESC
			LIMIT 1`, []string{"col0", "col1", "col3", "col5", "col6", "col7"}, nil},
		{`WITH
			with_273 (col_3485, col_3486, col_3487, col_3488, col_3489, col_3490, col_3491, col_3492, col_3493)
				AS (
					SELECT
						(-6623365040095722935)  AS col_3485,
						false AS col_3486,
						tab_1229.col1 AS col_3487,
						tab_1229.col6 AS col_3488,
						'1993-06-15'  AS col_3489,
						'rV'  AS col_3490,
						tab_1229.col4 AS col_3491,
						B'0110011001100' AS col_3492,
						tab_1229.col0 AS col_3493
					FROM
						(
							SELECT
								tab_1222.col5 AS col_3477,
								tab_1222.col4 AS col_3478,
								max(tab_1222.col3 )  AS col_3479,
								stddev(tab_1222.col5 )  AS col_3480
							FROM
								defaultdb.public.table1 AS tab_1222
							WHERE
								false
							GROUP BY
								tab_1222.col0, tab_1222.col4, tab_1222.col5, tab_1222.col3
							HAVING
								inet_same_family(((
									SELECT
										SET_MASKLEN('4ac:ded4:393a:a371:7690:9d0f:4817:3371/43', TAB_1226.COL5) AS  COL_3476
									FROM
										DEFAULTDB.PUBLIC.TABLE1 AS  TAB_1223
									RIGHT JOIN
										(
											SELECT
												TAB_1222.COL1   AS  COL_3470
											,   TAB_1222.COL0   AS  COL_3471
											,   TAB_1222.COL3   AS  COL_3472
											LIMIT 5
										)   AS  TAB_1224
									(COL_3473, COL_3474, COL_3475)
									JOIN
										DEFAULTDB.PUBLIC
									.TABLE0 AS  TAB_1225
									RIGHT JOIN
										DEFAULTDB.PUBLIC
									.TABLE1 AS  TAB_1226
									ON
										NULL
									FULL JOIN
										DEFAULTDB.PUBLIC
									.TABLE1 AS  TAB_1227
									ON
										FALSE
									ON  SIMILAR_TO_ESCAPE(TAB_1222.COL6, NULL, TAB_1222.COL6)
									ON  INET_CONTAINS_OR_EQUALS(
											SET_MASKLEN('198f:60f5:287a:8163:c091:2a95:afdc:ae8b/108',
												(-4475677368810664623)),
											'6d38:61ce:1af7:9283:cf0d:beb2:23e0:d7f/109')
									ORDER BY
										TAB_1226.COL7   DESC
									LIMIT 1
								)  - tab_1222.col5 )  , NULL )
						)
							AS tab_1228 (col_3481, col_3482, col_3483, col_3484),
						defaultdb.public.table1 AS tab_1229,
						defaultdb.public.table1 AS tab_1230
				)
		SELECT
			'c'  AS col_3494,
			e'\x00'  AS col_3495,
			tab_1232.col2 AS col_3496,
			tab_1232.col3 AS col_3497,
			3.4028234663852886e+38  AS col_3498,
			NULL AS col_3499
		FROM
			defaultdb.public.table1 AS tab_1231,
			defaultdb.public.table0 AS tab_1232,
			with_273,
			defaultdb.public.table0 AS tab_1233
		WHERE
			with_273.col_3486
		ORDER BY
			tab_1233.col0 DESC, tab_1232.col4 ASC, with_273.col_3490 ASC
		LIMIT
			23 ;`, []string{"col0", "col1", "col2", "col3", "col4", "col5", "col6", "col7", "col_3486", "col_3490"}, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.sql,  func(t *testing.T) {
			referredCols, err := func() (ReferredCols, error) {
				return ColNamesInSelect(tc.sql)
			}()
			if err.Error() != tc.err.Error() {
				t.Errorf("Expect %s, got %s", tc.err, err)
			}
			cols := referredCols.ToList()
			sort.Strings(tc.cols)
			for i, v := range cols {
				if v != tc.cols[i] {
					t.Errorf("Expect %s, got %s", tc.cols, cols)
					break
				}
			}
		})
	}
}
