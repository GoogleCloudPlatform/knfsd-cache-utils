/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strings"
	"text/template"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-fsidd/log"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed schema.sql
var tableSchema string

type DB interface {
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
}

type DBWrapper struct {
	dialer *cloudsqlconn.Dialer
	db     DB
}

func (w *DBWrapper) Close() {
	if w.db != nil {
		w.db.Close()
	}
	if w.dialer != nil {
		w.dialer.Close()
	}
}

func (w *DBWrapper) BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error {
	return w.db.BeginTxFunc(ctx, txOptions, f)
}

func (w *DBWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return w.db.Exec(ctx, sql, arguments...)
}

func (w *DBWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return w.db.QueryRow(ctx, sql, args...)
}

func connect(ctx context.Context, config DatabaseConfig) (DB, error) {
	pgConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, err
	}

	dialer, err := newDialer(ctx, config)
	if err != nil {
		return nil, err
	}

	pgConfig.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// ignore the host (addr) requested by pgx and instead use the cloud SQL instance
		return dialer.Dial(ctx, config.Instance)
	}

	log.Debug.Print("Creating pgxpool")
	db, err := pgxpool.ConnectConfig(ctx, pgConfig)
	if err != nil {
		dialer.Close()
		return nil, err
	}

	return &DBWrapper{dialer, db}, err
}

func newDialer(ctx context.Context, config DatabaseConfig) (*cloudsqlconn.Dialer, error) {
	var dialOptions []cloudsqlconn.DialOption
	var options []cloudsqlconn.Option

	if config.IAMAuth {
		options = append(options, cloudsqlconn.WithIAMAuthN())
	}

	if config.PrivateIP {
		dialOptions = append(dialOptions, cloudsqlconn.WithPrivateIP())
	} else {
		dialOptions = append(dialOptions, cloudsqlconn.WithPublicIP())
	}

	options = append(options, cloudsqlconn.WithDefaultDialOptions(dialOptions...))

	log.Debug.Print("Creating Cloud SQL dialer")
	dialer, err := cloudsqlconn.NewDialer(ctx, options...)
	if err != nil {
		return nil, err
	}

	log.Debug.Print("Warming up Cloud SQL dialer")
	err = dialer.Warmup(ctx, config.Instance)
	if err != nil {
		dialer.Close()
		return nil, err
	}

	return dialer, nil
}

type FSIDSource struct {
	db        DB
	tableName string
}

func (s FSIDSource) CreateTable(ctx context.Context) error {
	log.Debug.Printf("creating table \"%s\"", s.tableName)

	t, err := template.New("schema").Parse(tableSchema)
	if err != nil {
		return err
	}

	w := &strings.Builder{}
	err = t.Execute(w, s.tableName)
	if err != nil {
		return err
	}

	sql := w.String()
	return withRetry(ctx, func() error {
		_, err = s.db.Exec(ctx, sql)
		return err
	})
}

func (s FSIDSource) GetFSID(ctx context.Context, path string) (int32, error) {
	var fsid int32
	sql := fmt.Sprintf("SELECT fsid FROM \"%s\" WHERE path = $1", s.tableName)
	row := s.db.QueryRow(ctx, sql, path)
	err := row.Scan(&fsid)
	return fsid, err
}

func (s FSIDSource) AllocateFSID(ctx context.Context, path string) (int32, error) {
	var fsid int32
	sql := fmt.Sprintf("INSERT INTO \"%s\" (path) VALUES ($1) RETURNING fsid", s.tableName)
	row := s.db.QueryRow(ctx, sql, path)
	err := row.Scan(&fsid)
	return fsid, err
}

func (s FSIDSource) GetPath(ctx context.Context, fsid int32) (string, error) {
	var path string
	sql := fmt.Sprintf("SELECT path FROM \"%s\" WHERE fsid = $1", s.tableName)
	row := s.db.QueryRow(ctx, sql, fsid)
	err := row.Scan(&path)
	return path, err
}

func IsConflict(err error) bool {
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		// unique constraint violation
		return pgerr.Code == "23505"
	} else {
		return false
	}
}
