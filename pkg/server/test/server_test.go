package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/openfga/openfga/cmd/util"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"
	"github.com/openfga/openfga/pkg/storage/mysql"
	"github.com/openfga/openfga/pkg/storage/postgres"
	"github.com/openfga/openfga/pkg/storage/sqlcommon"

	storagefixtures "github.com/openfga/openfga/pkg/testfixtures/storage"
)

func TestServerWithPostgresDatastore(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})
	_, ds, _ := util.MustBootstrapDatastore(t, "postgres")

	s := server.MustNewServerWithOpts(server.WithDatastore(ds))
	t.Cleanup(s.Close)

	RunAllTests(t, ds, s)
}

func TestServerWithPostgresDatastoreAndExplicitCredentials(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})
	testDatastore := storagefixtures.RunDatastoreTestContainer(t, "postgres")

	uri := testDatastore.GetConnectionURI(false)
	ds, err := postgres.New(
		uri,
		sqlcommon.NewConfig(
			sqlcommon.WithUsername(testDatastore.GetUsername()),
			sqlcommon.WithPassword(testDatastore.GetPassword()),
		),
	)
	require.NoError(t, err)

	s := server.MustNewServerWithOpts(server.WithDatastore(ds))
	t.Cleanup(s.Close)

	RunAllTests(t, ds, s)
}

func TestServerWithMemoryDatastore(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})
	_, ds, _ := util.MustBootstrapDatastore(t, "memory")

	s := server.MustNewServerWithOpts(server.WithDatastore(ds))
	t.Cleanup(s.Close)

	RunAllTests(t, ds, s)
}

func TestServerWithMySQLDatastore(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})
	_, ds, _ := util.MustBootstrapDatastore(t, "mysql")
	s := server.MustNewServerWithOpts(server.WithDatastore(ds))
	t.Cleanup(s.Close)

	RunAllTests(t, ds, s)
}

func TestServerWithMySQLDatastoreAndExplicitCredentials(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t)
	})
	testDatastore := storagefixtures.RunDatastoreTestContainer(t, "mysql")

	uri := testDatastore.GetConnectionURI(false)
	ds, err := mysql.New(
		uri,
		sqlcommon.NewConfig(
			sqlcommon.WithUsername(testDatastore.GetUsername()),
			sqlcommon.WithPassword(testDatastore.GetPassword()),
		),
	)
	require.NoError(t, err)
	s := server.MustNewServerWithOpts(server.WithDatastore(ds))
	t.Cleanup(s.Close)

	RunAllTests(t, ds, s)
}

func BenchmarkOpenFGAServer(b *testing.B) {
	b.Cleanup(func() {
		goleak.VerifyNone(b,
			// https://github.com/uber-go/goleak/discussions/89
			goleak.IgnoreTopFunction("testing.(*B).run1"),
			goleak.IgnoreTopFunction("testing.(*B).doBench"),
		)
	})
	b.Run("BenchmarkPostgresDatastore", func(b *testing.B) {
		testDatastore := storagefixtures.RunDatastoreTestContainer(b, "postgres")

		uri := testDatastore.GetConnectionURI(true)
		ds, err := postgres.New(uri, sqlcommon.NewConfig())
		require.NoError(b, err)
		b.Cleanup(ds.Close)
		RunAllBenchmarks(b, ds)
	})

	b.Run("BenchmarkMemoryDatastore", func(b *testing.B) {
		ds := memory.New()
		b.Cleanup(ds.Close)
		RunAllBenchmarks(b, ds)
	})

	b.Run("BenchmarkMySQLDatastore", func(b *testing.B) {
		testDatastore := storagefixtures.RunDatastoreTestContainer(b, "mysql")

		uri := testDatastore.GetConnectionURI(true)
		ds, err := mysql.New(uri, sqlcommon.NewConfig())
		require.NoError(b, err)
		b.Cleanup(ds.Close)
		RunAllBenchmarks(b, ds)
	})
}