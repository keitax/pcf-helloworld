package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
)

func loadVcapServices() map[string]interface{} {
	vcapEnv, ok := os.LookupEnv("VCAP_SERVICES")
	if !ok {
		return map[string]interface{}{}
	}
	var vcapServices map[string]interface{}
	if err := json.Unmarshal([]byte(vcapEnv), &vcapServices); err != nil {
		logrus.Error(err)
		panic(err)
	}
	return vcapServices
}

func getDbName(credentials map[string]interface{}) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		credentials["username"].(string),
		credentials["password"].(string),
		credentials["hostname"].(string),
		credentials["port"].(string),
		credentials["name"].(string),
	)
}

func main() {
	vcapServices := loadVcapServices()

	clearDbCredentials := vcapServices["cleardb"].([]interface{})[0].(map[string]interface{})["credentials"].(map[string]interface{})
	logrus.Infof("dbname: %s", getDbName(clearDbCredentials))
	db, err := sql.Open("mysql", getDbName(clearDbCredentials))
	if err != nil {
		logrus.Fatal(err)
		panic(err)
	}
	defer db.Close()

	s := http.NewServeMux()
	s.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		q, ps, err := squirrel.Select("*").From("USERS").ToSql()
		if err != nil {
			panic(err)
		}
		rs, err := db.Query(q, ps...)
		if err != nil {
			logrus.Error(err)
			panic(err)
		}
		w.Header().Add("Content-Type", "text/plain")
		for rs.Next() {
			var id int
			var name string
			rs.Scan(&id, &name)
			fmt.Fprintf(w, "id: %d, name: %s\n", id, name)
		}
	})
	if err := http.ListenAndServe(":8080", s); err != nil {
		panic(err)
	}
}
