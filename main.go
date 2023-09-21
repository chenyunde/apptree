package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"time"
)

type treeNode struct {
	TreeNodeId   int        `json:"id"`
	TreeNodeName string     `json:"title"`
	FTreeNodeId  string     `json:"f_tree_node_id"`
	HasChild     bool       `json:"hasChild"`
	ChildNodes   []treeNode `json:"children"`
}

const (
	DRIVERNAME      = "mysql"
	USERNAME        = "root"
	PASSWORD        = "MyNewPass4!"
	PROTOCOL        = "tcp"
	IP              = "117.50.189.44"
	PORT            = "3306"
	DBNAME          = "tree"
	CONNMAXLIFETIME = time.Minute * 3
	MAXOPENCONNS    = 10
	MAXIDLECONNS    = 10
)

func main() {
	db := initMysql()
	defer db.Close()
	initTree(db)
	r := gin.Default()
	r.Use(Core())
	r.GET("/addnode", func(c *gin.Context) {
		node := treeNode{
			TreeNodeId:   2222,
			TreeNodeName: "test",
			FTreeNodeId:  "",
		}
		addNode(db, &node)
		c.JSON(http.StatusOK, gin.H{
			"status":       "200",
			"treeNodeId":   node.TreeNodeId,
			"treeNodeName": node.TreeNodeName,
			"fTreeNodeId":  node.FTreeNodeId,
		})
	})
	r.GET("/getnode", func(c *gin.Context) {
		nodes := initTree(db)
		c.JSON(http.StatusOK, gin.H{
			"status": "200",
			"nodes":  nodes,
		})
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}

// 解决跨域问题
func Core() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token")
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "True")
		//放行索引options
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}

// 链接数据库
func initMysql() *sql.DB {
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	dataSourceName := USERNAME + ":" + PASSWORD + "@" + PROTOCOL + "(" + IP + ":" + PORT + ")/" + DBNAME
	db, err := sql.Open(DRIVERNAME, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	// Important settings
	db.SetConnMaxLifetime(CONNMAXLIFETIME)
	db.SetMaxOpenConns(MAXOPENCONNS)
	db.SetMaxIdleConns(MAXIDLECONNS)
	return db
}

// 获取业务树
func initTree(db *sql.DB) []treeNode {
	var nodes []treeNode
	rows, err := db.Query("select * from tree where fTreeNodeId=\"\"")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var (
			treeNodeId   int
			treeNodeName string
			fTreeNodeId  string
		)
		if err := rows.Scan(&treeNodeId, &treeNodeName, &fTreeNodeId); err != nil {
			log.Fatal(err)
		}
		nodes = append(nodes, treeNode{
			TreeNodeId:   treeNodeId,
			TreeNodeName: treeNodeName,
			FTreeNodeId:  fTreeNodeId,
		})
	}
	getChildNodes(nodes, db)
	return nodes
}

// 递归获取子树
func getChildNodes(nodes []treeNode, db *sql.DB) {
	length := len(nodes)
	for i := 0; i < length; i++ {
		rows, err := db.Query("select * from tree where fTreeNodeId=?", nodes[i].TreeNodeId)
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var (
				treeNodeId   int
				treeNodeName string
				fTreeNodeId  string
			)
			if err := rows.Scan(&treeNodeId, &treeNodeName, &fTreeNodeId); err != nil {
				log.Fatal(err)
			}
			nodes[i].ChildNodes = append(nodes[i].ChildNodes, treeNode{
				TreeNodeId:   treeNodeId,
				TreeNodeName: treeNodeName,
				FTreeNodeId:  fTreeNodeId,
			})
		}
		if len(nodes[i].ChildNodes) > 0 {
			nodes[i].HasChild = true
		}
		getChildNodes(nodes[i].ChildNodes, db)
	}
}

// 增加一个应用
func addNode(db *sql.DB, node *treeNode) {
	_, err := db.Exec("insert into tree(treeNodeId,treeNodeName,fTreeNodeId) values(?,?,?);",
		node.TreeNodeId, node.TreeNodeName, node.FTreeNodeId)
	if err != nil {
		log.Fatal(err)
	}
}
