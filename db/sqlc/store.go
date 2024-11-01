package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provide all functoms to execute db quires and trabsactions
type Store struct {
	*Queries
	db *sql.DB
}


func NewStore(db *sql.DB)  *Store {
	return &Store{
		db: db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return nil
	}

	q := New(tx)

	err = fn(q)


	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}

// TransferTxParams context the input prameter of the transfer transaction
type TransferTxParams struct {
  FromAccountID int64 `json:"from_account_id"`
  ToAccountID int64 `json:"to_account_id"`
  Amount int64 `json:"amount"`
}

// TransferTxResult is the result of transfer transaction 
type TransferTxResult struct {
	Transfer Transfer `json:"transfer"`
	FromAccount Account `json:"from_account"`
  	ToAccount Account `json:"to_account"`
	FromEntry Entry `json:"from_entry"`
	ToEntry Entry `json:"to_entry"`
}



// TransferTx perform a money transfer from one account to another
/**
1. Create a transfer record 
2. Add account entries 
3. Update account's balance withing a single database transaction
*/

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error




		result.Transfer, err =q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})

		if err != nil {
			return err
		}


		// 2
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,
		})

		if err != nil{
			return nil
		}


		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})

		if err != nil{
			return nil
		}

		// 3
		// get account -> update it's balance


	

		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID: arg.FromAccountID,
			Amount: -arg.Amount,
		})

		if err != nil {
			return err
		}





		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID: arg.ToAccountID,
			Amount: arg.Amount,
		})

		if err != nil {
			return err
		}



		return nil
	})

	return result, err
}