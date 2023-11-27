package test

import (
	"context"
	"testing"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/openfga/pkg/server/commands"
	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/testutils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestListStores(t *testing.T, datastore storage.OpenFGADatastore) {
	ctx := context.Background()

	// clean up all stores from other tests
	getStoresQuery := commands.NewListStoresQuery(datastore)
	deleteCmd := commands.NewDeleteStoreCommand(datastore)
	deleteContinuationToken := ""
	for ok := true; ok; ok = deleteContinuationToken != "" {
		listStoresResponse, _ := getStoresQuery.Execute(ctx, &openfgav1.ListStoresRequest{
			ContinuationToken: deleteContinuationToken,
		})
		for _, store := range listStoresResponse.Stores {
			if _, err := deleteCmd.Execute(ctx, &openfgav1.DeleteStoreRequest{
				StoreId: store.Id,
			}); err != nil {
				t.Fatalf("failed cleaning stores with %v", err)
			}
		}
		deleteContinuationToken = listStoresResponse.ContinuationToken
	}

	// ensure there are actually no stores
	listStoresResponse, actualError := getStoresQuery.Execute(ctx, &openfgav1.ListStoresRequest{})
	require.NoError(t, actualError)
	require.Empty(t, listStoresResponse.Stores)

	// create two stores
	createStoreQuery := commands.NewCreateStoreCommand(datastore)
	firstStoreName := testutils.CreateRandomString(10)
	_, err := createStoreQuery.Execute(ctx, &openfgav1.CreateStoreRequest{Name: firstStoreName})
	require.NoError(t, err, "error creating store 1")

	secondStoreName := testutils.CreateRandomString(10)
	_, err = createStoreQuery.Execute(ctx, &openfgav1.CreateStoreRequest{Name: secondStoreName})
	require.NoError(t, err, "error creating store 2")
	// first page: 1st store
	listStoresResponse, actualError = getStoresQuery.Execute(ctx, &openfgav1.ListStoresRequest{
		PageSize:          wrapperspb.Int32(1),
		ContinuationToken: "",
	})
	require.NoError(t, actualError)
	require.Len(t, listStoresResponse.Stores, 1)
	require.Equal(t, firstStoreName, listStoresResponse.Stores[0].Name)
	require.NotEmpty(t, listStoresResponse.ContinuationToken)

	// first page: 2nd store
	secondListStoresResponse, actualError := getStoresQuery.Execute(ctx, &openfgav1.ListStoresRequest{
		PageSize:          wrapperspb.Int32(1),
		ContinuationToken: listStoresResponse.ContinuationToken,
	})
	require.NoError(t, actualError)
	require.Len(t, secondListStoresResponse.Stores, 1)
	require.Equal(t, secondStoreName, secondListStoresResponse.Stores[0].Name)
	// no token <=> no more results
	require.Empty(t, secondListStoresResponse.ContinuationToken)
}
