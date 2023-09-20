package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"time"
)

type treeNode struct {
	TreeNodeId   int
	TreeNodeName string
	FTreeNodeId  string
	ChildNodes   []treeNode
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
		fmt.Println(nodes)
		c.JSON(http.StatusOK, gin.H{
			"status": "200",
			"nodes":  nodes,
		})
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}

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

func getChildNodes(nodes []treeNode, db *sql.DB) {
	len := len(nodes)
	for i := 0; i < len; i++ {
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
		getChildNodes(nodes[i].ChildNodes, db)
	}
}
func addNode(db *sql.DB, node *treeNode) {
	_, err := db.Exec("insert into tree(treeNodeId,treeNodeName,fTreeNodeId) values(?,?,?);",
		node.TreeNodeId, node.TreeNodeName, node.FTreeNodeId)
	if err != nil {
		log.Fatal(err)
	}
}
