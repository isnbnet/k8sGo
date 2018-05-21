package main

import (
	"fmt"
	"time"
	//"net/http"
	"database/sql"
	"log"
	"strings"

	"github.com/DeanThompson/ginpprof"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-oci8"
)

type result struct {
	Status string
}

var server = "192.168.2.29"
var port = 1433
var user = "sa"
var password = "ynsa@0805"
var database = "ydw"

//连接字符串
var connString = fmt.Sprintf("server=%s;port%d;database=%s;user id=%s;password=%s", server, port, database, user, password)

//var Db, _ = sql.Open("oci8", "sync/sync@192.168.2.61/sync")

//var syncConnString=fmt.Sprintf("server=%s;port%d;database=%s;user id=%s;password=%s", server, port, database, user, password)

func main() {

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	var ret = new(result)
	// 	ret.Status = "ok"
	// 	if data, err := json.Marshal(ret); err == nil {
	// 		fmt.Fprintf(w, string(data))
	// 	}

	// })

	// http.HandleFunc("/proc", func(w http.ResponseWriter, r *http.Request) {
	// 	var ret = make(map[string][]string)
	// 	r.ParseForm()
	// 	ret = r.Form
	// 	if data, err := json.Marshal(ret); err == nil {
	// 		ExecProcedure()
	// 		fmt.Fprintf(w, string(data))
	// 	}

	// })

	// http.ListenAndServe(":8080", nil)

	// if err != nil {
	// 	log.Fatal(err)
	// 	panic("数据库连接失败")
	// } else {
	//defer Db.Close()
	//}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	ginpprof.Wrap(r)
	r.GET("/login", JWTLogin())

	authorize := r.Group("/", JWTAuth())
	{
		// authorize.GET("user", func(c *gin.Context) {
		//     claims := c.MustGet("claims").(*jwtauth.CustomClaims)
		//     fmt.Println(claims.ID)
		//     c.String(http.StatusOK, claims.Name)

		authorize.GET("/proc", Auth(ExecProcedure()))
		//	authorize.GET("/sync/updateinfo", Auth(UpdateSyncQtyToSyncDb()))

	}

	r.Run("0.0.0.0:80") // listen and serve on 0.0.0.0:8080

	// RedisTest()
}

// func UpdateSyncQtyToSyncDb() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var eq int
// 		var bc = c.Query("b")
// 		var sq, err = strconv.Atoi(c.Query("sq"))
// 		eq, err = strconv.Atoi(c.Query("eq"))
// 		var sqlText = fmt.Sprintf(`begin update (
// 		select o.*
// 		from sc_online o
// 		inner join sc_branch b on o.id=b.id
// 		where b.branchcode='%s'
// 		) set errsyncqty=%d,NOTSYNCQTY=%d,lastonlinedate=sysdate;  end;
// 		`, bc, eq, sq)
// 		//fmt.Println(sqlText)

// 		rows, err := Db.Query(sqlText)

// 		if err != nil {
// 			Db, _ = sql.Open("oci8", "sync/sync@192.168.2.61/sync")
// 			rows, err = Db.Query(sqlText)
// 		}

// 		defer rows.Close()

// 		if err != nil {
// 			c.JSON(400, err)
// 		}
// 		c.String(200, "OK")

// 	}
// }

func ExecOracleProcedure() gin.HandlerFunc {
	//用户名/密码@实例名 如system/123456@orcl、sys/123456@orcl
	return func(c *gin.Context) {

		var p = c.Query("p")

		db, err := sql.Open("oci8", "sync/sync@192.168.2.61/sync")
		if err != nil {
			log.Fatal(err)
			panic("数据库连接失败")
		} else {
			defer db.Close()
			var sqlText = fmt.Sprintf(`begin
			%s('Y02032',9000);
			end;`, p)
			_, err := db.Query(sqlText)
			if err != nil {
				c.JSON(400, err)
			}

			// for rows.Next() {
			// 	var f2 int
			// 	var f1 int
			// 	rows.Scan(&f1, &f2)
			// 	println(f1, f2) // 3.14 foo
			// }
		}
	}
}

func ExecProcedure() gin.HandlerFunc {
	return func(c *gin.Context) {
		var p = c.Query("p")
		var parasIn = []string{}
		var parasOut = []string{}
		var (
			decl = ""
			out  = ""
		)

		for i := 0; i >= 0; i++ {
			var ps = c.Query(fmt.Sprintf("p%d", i))
			if ps == "" {
				break
			}
			var para1 = strings.Split(ps, "|")

			//返回值，直接存储过程里Select返回。不用参数数返回。省略以下代码
			// if para1[2] == "0" {
			// 	parasIn = append(parasIn, para1[0])
			// } else {
			// 	switch para1[1][0] {
			// 	case 'i':
			// 		decl = fmt.Sprintf("%v@%v int,", decl, para1[0])
			// 		out = fmt.Sprintf("%v@%v,", out, para1[0])
			// 	case 'a':
			// 		decl = fmt.Sprintf("%v@%v varchar(%v),", decl, para1[0], para1[1][1:])
			// 		out = fmt.Sprintf("%v@%v,", out, para1[0])
			// 	default:
			// 		fmt.Printf("支持参数 %s", para1[1][0])

			// 	}
			// 	str1 := fmt.Sprintf("@%v", para1[0])
			// 	fmt.Print(str1)
			// 	parasOut = append(parasOut, fmt.Sprintf("@%v", para1[0]))
			// }

			parasIn = append(parasIn, para1[0])
		}

		// if decl != "" {
		// 	decl = "declare " + decl[0:len(decl)-1]
		// 	out = out[0 : len(out)-1]
		// }

		conn, err := sql.Open("mssql", connString)
		if err != nil {
			log.Fatal("Open Connection failed:", err.Error())
		}
		defer conn.Close()
		sqlExits := fmt.Sprintf("select 1 cnt from sys.objects where type='P' and name='%s'", p)
		m, _ := exec(conn, sqlExits)
		//存在存储过程过程就执行
		if len(m) > 0 {
			var in = strings.Join(parasIn, ",")
			sqlString := GetProcSql(p, decl, in, out, parasOut...)
			//fmt.Println(sqlString)
			m, _ = exec(conn, sqlString)
		}

		//fmt.Println(json.Marshal(m))

		if err != nil {
			log.Fatal("Query 失败！")
		}

		c.JSON(200, m)
	}
}

func exec(db *sql.DB, cmd string) ([]map[string]string, error) {
	var ret = make([]map[string]string, 0)
	rows, err := db.Query(cmd)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return ret, err
	}
	if cols == nil {
		return ret, err
	}
	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}
	fmt.Println()
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			fmt.Println(err)
			continue
		}
		c := make(map[string]string)
		for i := 0; i < len(cols); i++ {
			pval := vals[i].(*interface{})
			//fmt.Println((*pval).(type))
			switch v := (*pval).(type) {
			case nil:
				c[cols[i]] = fmt.Sprint("NULL")
			case bool:
				if v {
					c[cols[i]] = fmt.Sprint("1")
				} else {
					c[cols[i]] = fmt.Sprint("0")
				}
			case []byte:
				c[cols[i]] = string(v)

			case time.Time:
				c[cols[i]] = fmt.Sprint(v.Format("2006-01-02 15:04:05.999"))
			case int:
			case int64:
				c[cols[i]] = fmt.Sprintf("%d", v)
			default:
				fmt.Printf("不知道类型 %T", v)

			}
		}
		ret = append(ret, c)
	}
	if rows.Err() != nil {
		return ret, err
	}
	return ret, err
}

//proc is the proc name
//declare is the proc declare with the return values
//in is the params in
//out is the params out
//outparas is the select parameters
func GetProcSql(proc, declare, in, out string, outparas ...string) string {
	_sql := fmt.Sprintf("%v;exec %v %v ", declare, proc, in)
	var outparam string
	for _, out := range outparas {
		outparam = fmt.Sprintf("%v,%v=%v OUTPUT", outparam, out, out)
	}
	outparam = fmt.Sprintf("%v;", outparam)
	if out != "" {
		_sql = fmt.Sprintf("%v%vselect %v;", _sql, outparam, out)
	} else {
		_sql = fmt.Sprintf("%v%v", _sql, outparam)
	}
	return _sql

}
