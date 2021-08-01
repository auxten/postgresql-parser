package main

import (
	"fmt"
	"log"

	"github.com/auxten/postgresql-parser/pkg/sql/parser"
	"github.com/auxten/postgresql-parser/pkg/sql/sem/tree"
)

func main() {
	sql := "SELECT to_date('20201230', 'YYYYMMDD') AS data_dt, COALESCE(t4.party_id, t.sign_org) AS org_id, t.agmt_id, " +
		"t1.fcy_spec_acct_id_type AS fcy_spec_acct_id_type, t.agmt_mdfr, t.categ_cd, COALESCE(NULL, '') AS retail_ind, " +
		"t.item_id, CASE WHEN t.categ_cd IN ('1013', '1021') THEN '1101' WHEN t.categ_cd IN ('4009',) THEN '1102' WHEN " +
		"t.categ_cd IN ('4013',) THEN '1201' WHEN (t.categ_cd IN ('4010',)) AND (t3.inform_deposit_categ IN ('SLA', 'EDA')) " +
		"THEN '1202' WHEN (t.categ_cd IN ('4012',)) AND (t3.inform_deposit_categ NOT IN ('NRI1', 'NRI7', 'IMM7', 'REN7')) " +
		"THEN '1301' WHEN (t.categ_cd IN ('4012',)) AND (t3.inform_deposit_categ IN ('NRI1', 'NRI7')) THEN '1302' WHEN " +
		"(t.categ_cd IN ('4012',)) AND (t3.inform_deposit_categ IN ('IMM7', 'REN7')) THEN '1303' ELSE '9999' END AS prod_cd, " +
		"t3.inform_deposit_categ AS sub_prod_cd, t.party_id AS cust_id, t.categ_cd AS acct_categ_cd, t.agmt_categ_cd AS " +
		"acct_type_cd, CASE WHEN t.categ_cd = '1604' THEN '2' ELSE '1' END AS acct_stat_cd, CASE WHEN " +
		"t1.sleep_acct_ind = 'Y' THEN '1' ELSE '0' END AS dormancy_ind, CASE WHEN t1.dep_exchg_ind = 'Y' THEN '1' " +
		"ELSE '0' END AS dep_exchg_ind, COALESCE(t11.intr, 0.00) AS mature_intr, " +
		"COALESCE(t.sign_dt, to_date('${NULLDATE}', 'YYYYMMDD')) AS open_acct_dt, " +
		"COALESCE(t.st_int_dt, to_date('${NULLDATE}', 'YYYYMMDD')) AS st_int_dt, " +
		"COALESCE(t.mature_dt, to_date('${MAXDATE}', 'YYYYMMDD')) AS mature_dt, CASE WHEN (t.src_sys = 'S04_ACCT_CLOSED') " +
		"AND (t.agmt_stat_cd = 'XH') THEN t.close_dt ELSE to_date('${MAXDATE}', 'YYYYMMDD') " +
		"END AS close_acct_dt, t9.agenter_nm AS agenter_nm, CASE WHEN t9.agenter_ident_info_categ_cd = 'CD_018' " +
		"THEN t9.agenter_ident_info_categ_cd ELSE '' END AS agenter_cert_type_cd, t9.agenter_ident_info_content AS " +
		"agenter_cert_id, t9.agenter_nationality_cd AS agenter_nationality, t9.agenter_tel AS agenter_tel, " +
		"t9.agent_open_acct_verify_situati AS agenter_open_acct_verify_rslt, t.ccy_cd, t.open_acct_amt, " +
		"COALESCE(substr(t5.tid, 1, 3), '') AS ftz_actype, CASE WHEN (t5.tid IS NOT NULL) OR (t5.tid != '') " +
		"THEN '1' ELSE '0' END AS ftz_act_ind, CASE WHEN t.categ_cd = '4012' THEN 'D' ELSE " +
		"COALESCE(t6.term_unit_cd, '') END AS term_type_cd, CASE WHEN t.categ_cd = '4012' " +
		"THEN to_number(substr(sub_prod_cd, 4, 1), '9') ELSE COALESCE(t6.term, 0) END AS " +
		"deposit_periods, COALESCE(CASE WHEN ((((prod_cd = '1101') OR (t.item_id IN ('14002', '15002', '16002'))) " +
		"OR (COALESCE(t.sign_dt, to_date('${NULLDATE}', 'YYYYMMDD')) = to_date('${NULLDATE}', 'YYYYMMDD'))) OR " +
		"(COALESCE(t.mature_dt, to_date('${MAXDATE}', 'YYYYMMDD')) = to_date('${NULLDATE}', 'YYYYMMDD'))) OR " +
		"(COALESCE(t.mature_dt, to_date('${MAXDATE}', 'YYYYMMDD')) = to_date('${MAXDATE}', 'YYYYMMDD')) THEN " +
		"0 ELSE COALESCE(t.mature_dt, to_date('${MAXDATE}', 'YYYYMMDD')) - t.st_int_dt END, 0) AS term_days, " +
		"CASE WHEN prod_cd = '1101' THEN '' WHEN t.item_id IN ('06003', '011', '01014', '01015', '01016', '01017', '099')" +
		" THEN 'M' WHEN t.item_id IN ('4002', '5002', '6002') THEN 'D' ELSE (CASE WHEN t6.term_unit_cd IS NOT NULL " +
		"THEN t6.term_unit_cd ELSE (CASE WHEN term_days < 7 THEN '' WHEN (term_days >= 7) AND (term_days < 28) THEN " +
		"'D' ELSE 'M' END) END) END AS adj_term_type_cd, CASE WHEN prod_cd = '1101' THEN 0 WHEN t.item_id = '1006003'" +
		" THEN 60 WHEN t.item_id = '10011' THEN 3 WHEN t.item_id IN ('1001014', '1001015', '1001016', '1099') " +
		"THEN 12 WHEN t.item_id = '1001017' THEN 24 ELSE (CASE WHEN deposit_periods > 0 THEN deposit_periods ELSE" +
		" (CASE WHEN term_days < 7 THEN 0 WHEN (term_days >= 7) AND (term_days < 28) THEN 7 WHEN (term_days >= 28) " +
		"AND (term_days <= 31) THEN 1 WHEN (term_days > 31) AND (term_days <= 92) THEN 3 WHEN (term_days > 92) AND " +
		"(term_days <= 184) THEN 6 WHEN (term_days > 184) AND (term_days <= 366) THEN 12 WHEN (term_days > 366) AND " +
		"(term_days <= 731) THEN 24 WHEN (term_days > 731) AND (term_days <= 1096) THEN 36 WHEN term_days > 1096 " +
		"THEN 60 END) END) END AS adj_deposit_periods, COALESCE(NULL, '') AS product_code, COALESCE(NULL, '') AS " +
		"lmt_lnk_ind, COALESCE(t.cur_bal, 0.00) AS open_cleared_bal, t7.assoc_agmt_id AS limit_ref, " +
		"COALESCE(t8.cash_pool_group, '') AS cash_pool_group, COALESCE(t10.medium_id, '') AS card_id FROM " +
		"agmt_item_temp AS t LEFT JOIN pviewdb.t03_acct AS t1 ON t.agmt_id = t1.agmt_id LEFT JOIN " +
		"pviewdb.t03_inform_dep_acct AS t3 ON ((t3.agmt_id = t.agmt_id) AND (t3.st_dt <= to_date('20201230', 'YYYYMMDD')))" +
		" AND (t3.end_dt > to_date('20201230', 'YYYYMMDD')) LEFT JOIN t03_agmt_pty_rela_h_temp AS t4 ON" +
		" t4.agmt_id = t.agmt_id LEFT JOIN s04_zmq_acc_cur AS t5 ON t5.customer = t.party_id LEFT JOIN acct_term_temp" +
		" AS t6 ON t.agmt_id = t6.agmt_id LEFT JOIN t03_agmt_rela_h_temp AS t7 ON t.agmt_id = t7.agmt_id " +
		"LEFT JOIN agmt_cash_pool_temp AS t8 ON t.agmt_id = t8.tid LEFT JOIN pviewdb.t03_agmt_agent_h AS t9" +
		" ON ((t.agmt_id = t9.agmt_id) AND (t9.st_dt <= to_date('20201230', 'YYYYMMDD'))) AND" +
		" (t9.end_dt > to_date('20201230', 'YYYYMMDD')) LEFT JOIN pviewdb.t03_agmt_medium_rela_h " +
		"AS t10 ON (((t.agmt_id = t10.agmt_id) AND (t10.st_dt <= to_date('20201230', 'YYYYMMDD'))) " +
		"AND (t10.end_dt > to_date('20201230', 'YYYYMMDD'))) AND (t10.agmt_medium_rela_type_cd = '2')" +
		" LEFT JOIN pviewdb.t03_agmt_int_h AS t11 ON ((((t.agmt_id = t11.agmt_id) AND (t.agmt_mdfr = t11.agmt_mdfr)) " +
		"AND (t11.st_dt <= to_date('20201230', 'YYYYMMDD'))) AND (t11.end_dt > to_date('20201230', 'YYYYMMDD'))) " +
		"AND (t11.int_type_cd = '7');"

	if stmts, err := parser.Parse(sql); err != nil{
		log.Fatal(err)
	} else {
		for _, stmt := range stmts {
			fmt.Printf("Pretty SQL:\n%s", tree.Pretty(stmt.AST))
		}
	}

}
