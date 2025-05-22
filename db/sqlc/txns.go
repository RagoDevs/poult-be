package db

import (
	"context"
	"time"
)

const CHICKEN_PURCHASE = "chicken_purchase"
const CHICKEN_SALE = "chicken_sale"

type TxnRequest struct {
	Type           string          `json:"type" validate:"required,oneof=expense income"`
	Category       string          `json:"category" validate:"required,oneof=food medicine chicken tools other"`
	Amount         int32           `json:"amount" validate:"required,gt=0"`
	Date           time.Time       `json:"date" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Quantity       *int32          `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	ChickenType    *string         `json:"chickenType,omitempty" validate:"omitempty,oneof=chicks hen cock"`
	BulkQuantities *BulkQuantities `json:"bulkQuantities,omitempty"`
}

type BulkQuantities struct {
	Hen    int32 `json:"hen" validate:"gte=0"`
	Cock   int32 `json:"cock" validate:"gte=0"`
	Chicks int32 `json:"chicks" validate:"gte=0"`
}

func (store *SQLStore) TxnCreateTransaction(ctx context.Context, args TxnRequest) error {

	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := New(tx)

	if args.Type == "expense" {

		if args.Category != "chicken" {

			ctg, err := qtx.GetCategoryByName(ctx, args.Category)
			if err != nil {
				return err
			}

			err = qtx.CreateTransaction(ctx, CreateTransactionParams{
				Type:        TransactionType(args.Type),
				CategoryID:  ctg.ID,
				Amount:      args.Amount,
				Date:        args.Date,
				Description: args.Description,
			})
			if err != nil {
				return err
			}

		} else if args.Category == "chicken" {
			
			ctg, err := qtx.GetCategoryByName(ctx, CHICKEN_PURCHASE)
			if err != nil {
				return err
			}

			err = qtx.CreateTransaction(ctx, CreateTransactionParams{
				Type:        TransactionType(args.Type),
				CategoryID:  ctg.ID,
				Amount:      args.Amount,
				Date:        args.Date,
				Description: args.Description,
			})
			if err != nil {
				return err
			}

			if args.ChickenType != nil {

				chicken, err := qtx.GetChickenByType(ctx, ChickenType(*args.ChickenType))
				if err != nil {
					return err
				}

				err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
					ID:       chicken.ID,
					Quantity: chicken.Quantity + *args.Quantity,
				})
				if err != nil {
					return err
				}

				err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
					ChickenType:    chicken.Type,
					QuantityChange: *args.Quantity,
					Reason:         ReasonType("purchase"),
				})

				if err != nil {
					return err
				}

			} else if args.BulkQuantities != nil {

				if args.BulkQuantities.Hen > 0 {
					chicken, err := qtx.GetChickenByType(ctx, ChickenType("hen"))
					if err != nil {
						return err
					}

					err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
						ID:       chicken.ID,
						Quantity: chicken.Quantity + args.BulkQuantities.Hen,
					})
					if err != nil {
						return err
					}

					err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
						ChickenType:    chicken.Type,
						QuantityChange: args.BulkQuantities.Hen,
						Reason:         ReasonType("purchase"),
					})
					if err != nil {
						return err
					}
				}

				if args.BulkQuantities.Cock > 0 {
					chicken, err := qtx.GetChickenByType(ctx, ChickenType("cock"))
					if err != nil {
						return err
					}

					err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
						ID:       chicken.ID,
						Quantity: chicken.Quantity + args.BulkQuantities.Cock,
					})
					if err != nil {
						return err
					}

					err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
						ChickenType:    chicken.Type,
						QuantityChange: args.BulkQuantities.Cock,
						Reason:         ReasonType("purchase"),
					})
					if err != nil {
						return err
					}
				}

				if args.BulkQuantities.Chicks > 0 {
					chicken, err := qtx.GetChickenByType(ctx, ChickenType("chicks"))
					if err != nil {
						return err
					}

					err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
						ID:       chicken.ID,
						Quantity: chicken.Quantity + args.BulkQuantities.Chicks,
					})
					if err != nil {
						return err
					}

					err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
						ChickenType:    chicken.Type,
						QuantityChange: args.BulkQuantities.Chicks,
						Reason:         ReasonType("purchase"),
					})
					if err != nil {
						return err
					}
				}

			}

		}

	} else if args.Type == "income" {
		
		ctg, err := qtx.GetCategoryByName(ctx, CHICKEN_SALE)
		if err != nil {
			return err
		}

		err = qtx.CreateTransaction(ctx, CreateTransactionParams{
			Type:        TransactionType(args.Type),
			CategoryID:  ctg.ID,
			Amount:      args.Amount,
			Date:        args.Date,
			Description: args.Description,
		})
		if err != nil {
			return err
		}

		if args.ChickenType != nil {

			chicken, err := qtx.GetChickenByType(ctx, ChickenType(*args.ChickenType))
			if err != nil {
				return err
			}

			err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
				ID:       chicken.ID,
				Quantity: chicken.Quantity - *args.Quantity,
			})
			if err != nil {
				return err
			}

			err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
				ChickenType:    chicken.Type,
				QuantityChange: -(*args.Quantity),
				Reason:         ReasonType("sale"),
			})
			if err != nil {
				return err
			}

		} else if args.BulkQuantities != nil {
			
			if args.BulkQuantities.Hen > 0 {
				chicken, err := qtx.GetChickenByType(ctx, ChickenType("hen"))
				if err != nil {
					return err
				}

				err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
					ID:       chicken.ID,
					Quantity: chicken.Quantity - args.BulkQuantities.Hen,
				})
				if err != nil {
					return err
				}

				err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
					ChickenType:    chicken.Type,
					QuantityChange: -args.BulkQuantities.Hen,
					Reason:         ReasonType("sale"),
				})
				if err != nil {
					return err
				}
			}

			if args.BulkQuantities.Cock > 0 {
				chicken, err := qtx.GetChickenByType(ctx, ChickenType("cock"))
				if err != nil {
					return err
				}

				err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
					ID:       chicken.ID,
					Quantity: chicken.Quantity - args.BulkQuantities.Cock,
				})
				if err != nil {
					return err
				}

				err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
					ChickenType:    chicken.Type,
					QuantityChange: -args.BulkQuantities.Cock,
					Reason:         ReasonType("sale"),
				})
				if err != nil {
					return err
				}
			}

			if args.BulkQuantities.Chicks > 0 {
				chicken, err := qtx.GetChickenByType(ctx, ChickenType("chicks"))
				if err != nil {
					return err
				}

				err = qtx.UpdateChickenById(ctx, UpdateChickenByIdParams{
					ID:       chicken.ID,
					Quantity: chicken.Quantity - args.BulkQuantities.Chicks,
				})
				if err != nil {
					return err
				}

				err = qtx.InsertChickenHistory(ctx, InsertChickenHistoryParams{
					ChickenType:    chicken.Type,
					QuantityChange: -args.BulkQuantities.Chicks,
					Reason:         ReasonType("sale"),
				})
				if err != nil {
					return err
				}
			}

		}
	}

	return tx.Commit()

}
