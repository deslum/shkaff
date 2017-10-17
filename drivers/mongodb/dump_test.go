package mongo

import (
	"fmt"
	"testing"
)

func TestMongoParams_ParamsToString(t *testing.T) {
	type fields struct {
		host                   string
		port                   int
		login                  string
		password               string
		ipv6                   bool
		database               string
		collection             string
		gzip                   bool
		parallelCollectionsNum int
	}
	tests := []struct {
		name              string
		fields            fields
		wantCommandString string
	}{
		// {
		// 	name: "Full",
		// 	fields: fields{
		// 		host:       "127.0.0.1",
		// 		port:       27017,
		// 		ipv6:       true,
		// 		login:      "test",
		// 		password:   "test",
		// 		database:   "dbtest",
		// 		collection: "colltest",
		// 	},
		// 	wantCommandString: "mongodump --host 127.0.0.1 --port 27017 --login test --password test --ipv6 --database dbtest --collection colltest",
		// },
		// {
		// 	name: "Bases empty",
		// 	fields: fields{
		// 		host:     "127.0.0.1",
		// 		port:     27017,
		// 		ipv6:     true,
		// 		login:    "test",
		// 		password: "test",
		// 	},
		// 	wantCommandString: "mongodump --host 127.0.0.1 --port 27017 --login test --password test --ipv6",
		// },
		// {
		// 	name: "IPv6 empty",
		// 	fields: fields{
		// 		host:     "127.0.0.1",
		// 		port:     27017,
		// 		ipv6:     false,
		// 		login:    "test",
		// 		password: "test",
		// 		database: "admin",
		// 	},
		// 	wantCommandString: "mongodump --host 127.0.0.1 --port 27017 --login test --password test --database dbtest --collection colltest",
		// },
		{
			name: "Auth empty",
			fields: fields{
				host:                   "127.0.0.1",
				port:                   27017,
				ipv6:                   false,
				gzip:                   true,
				database:               "admin",
				parallelCollectionsNum: 10,
			},
			wantCommandString: "mongodump --host 127.0.0.1 --port 27017 --database dbtest --collection colltest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MongoParams{
				host:       tt.fields.host,
				port:       tt.fields.port,
				login:      tt.fields.login,
				password:   tt.fields.password,
				ipv6:       tt.fields.ipv6,
				database:   tt.fields.database,
				collection: tt.fields.collection,
				gzip:       tt.fields.gzip,
				parallelCollectionsNum: tt.fields.parallelCollectionsNum,
			}
			// if gotCommandString := mp.ParamsToString(); gotCommandString != tt.wantCommandString {
			// 	t.Errorf("MongoParams.ParamsToString() = %v, want %v", gotCommandString, tt.wantCommandString)
			// }
			fmt.Println(mp.ParamsToString())
			mp.Dump()
		})
	}
}

