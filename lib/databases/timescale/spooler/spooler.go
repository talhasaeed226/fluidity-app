// Copyright 2022 Fluidity Money. All rights reserved. Use of this
// source code is governed by a GPL-style license that can be found in the
// LICENSE.md file.

package spooler

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/fluidity-money/fluidity-app/lib/log"
	"github.com/fluidity-money/fluidity-app/lib/timescale"
	"github.com/fluidity-money/fluidity-app/lib/types/misc"
	"github.com/fluidity-money/fluidity-app/lib/types/network"
	token_details "github.com/fluidity-money/fluidity-app/lib/types/token-details"
	"github.com/fluidity-money/fluidity-app/lib/types/worker"
)

const (
	// Context to use for logging
	Context = "TIMESCALE/SPOOLER"

	// TablePendingWinners to write pending winners for the
	// ethereum spooler to use
	TablePendingWinners = "ethereum_pending_winners"
)

func InsertPendingWinners(winner worker.EthereumWinnerAnnouncement) {
	timescaleClient := timescale.Client()

	var (
		tokenDetails = winner.TokenDetails

		tokenShortName     = tokenDetails.TokenShortName
		tokenDecimals      = tokenDetails.TokenDecimals
		network_           = winner.Network
		hash               = winner.TransactionHash
		blockNumber        = winner.BlockNumber
		senderAddress      = winner.FromAddress
		senderWinAmount    = winner.FromWinAmount
		recipientAddress   = winner.ToAddress
		recipientWinAmount = winner.ToWinAmount
	)

	statementText := fmt.Sprintf(
		`INSERT INTO %s (
			token_short_name,
			token_decimals,
			transaction_hash,
			address,
			win_amount,
			block_number,
			network,
			reward_type
		)

		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		);`,

		TablePendingWinners,
	)

	// insert the sender's winnings
	_, err := timescaleClient.Exec(
		statementText,
		tokenShortName,
		tokenDecimals,
		hash,
		senderAddress,
		senderWinAmount,
		blockNumber,
		network_,
		"send",
	)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to insert pending winner %+v!",
				winner,
			)

			k.Payload = err
		})
	}

	// insert the recipient's winnings
	_, err = timescaleClient.Exec(
		statementText,
		tokenShortName,
		tokenDecimals,
		hash,
		recipientAddress,
		recipientWinAmount,
		blockNumber,
		network_,
		"receive",
	)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to insert pending winner %+v!",
				winner,
			)

			k.Payload = err
		})
	}
}

func UnpaidWinningsForToken(network_ network.BlockchainNetwork, token token_details.TokenDetails) *big.Int {
	timescaleClient := timescale.Client()

	statementText := fmt.Sprintf(
		`SELECT
			SUM(win_amount)

		FROM %s
		WHERE
			network = $1
			AND token_short_name = $2
			AND reward_sent = false
		`,

		TablePendingWinners,
	)

	row := timescaleClient.QueryRow(
		statementText,
		network_,
		token.TokenShortName,
	)

	var total misc.BigInt

	err := row.Scan(&total)

	if err == sql.ErrNoRows {
		return big.NewInt(0)
	}

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to get unpaid winnings for token %s!",
				token.TokenShortName,
			)

			k.Payload = err
		})
	}

	return &total.Int
}

func GetAndRemoveRewardsForToken(network_ network.BlockchainNetwork, token token_details.TokenDetails) []worker.EthereumReward {
	timescaleClient := timescale.Client()

	shortName := token.TokenShortName

	statementText := fmt.Sprintf(
		`UPDATE %s
			SET reward_sent = true
		WHERE
			reward_sent = false
			AND network = $1
			AND token_short_name = $2
		RETURNING
			network,
			token_short_name,
			token_decimals,
			transaction_hash,
			address,
			win_amount,
			block_number
		;`,

		TablePendingWinners,
	)

	rows, err := timescaleClient.Query(
		statementText,
		network_,
		shortName,
	)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to fetch and mark winners as sent for token %s!",
				shortName,
			)

			k.Payload = err
		})
	}

	defer rows.Close()

	winners := make([]worker.EthereumReward, 0)

	for rows.Next() {
		var (
			winner worker.EthereumReward
		)

		err := rows.Scan(
			&winner.Network,
			&winner.TokenDetails.TokenShortName,
			&winner.TokenDetails.TokenDecimals,
			&winner.TransactionHash,
			&winner.Winner,
			&winner.WinAmount,
			&winner.BlockNumber,
		)

		if err != nil {
			log.Fatal(func(k *log.Log) {
				k.Context = Context
				k.Message = "Failed to scan a row of the pending winners!"
				k.Payload = err
			})
		}

		winners = append(winners, winner)
	}

	return winners
}

func GetPendingRewardsForAddress(network_ network.BlockchainNetwork, address string) []worker.EthereumReward {
	timescaleClient := timescale.Client()

	statementText := fmt.Sprintf(
		`SELECT
			token_short_name,
			token_decimals,
			transaction_hash,
			address,
			win_amount,
			block_number
		FROM %s
		WHERE
			reward_sent = false
			AND network = $1
			AND address = $2
		;
		`,

		TablePendingWinners,
	)

	rows, err := timescaleClient.Query(
		statementText,
		network_,
		address,
	)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to fetch unpaid winnings for user %s!",
				address,
			)

			k.Payload = err
		})
	}

	defer rows.Close()

	winners := make([]worker.EthereumReward, 0)

	for rows.Next() {
		var (
			winner worker.EthereumReward
		)

		err := rows.Scan(
			&winner.TokenDetails.TokenShortName,
			&winner.TokenDetails.TokenDecimals,
			&winner.TransactionHash,
			&winner.Winner,
			&winner.WinAmount,
			&winner.BlockNumber,
		)

		if err != nil {
			log.Fatal(func(k *log.Log) {
				k.Context = Context
				k.Message = "Failed to scan a row of pending winners!"
				k.Payload = err
			})
		}

		winners = append(winners, winner)
	}

	return winners
}

func RemovePendingWinnings(network_ network.BlockchainNetwork, token token_details.TokenDetails, address string, startBlock, endBlock *misc.BigInt) {
	timescaleClient := timescale.Client()

	statementText := fmt.Sprintf(
		`UPDATE %s

		SET
			reward_sent = true
		WHERE
			winner = $1
			AND network_ = $2
			AND token_short_name = $3
			AND block_number >= $4
			AND block_number <= $5
		;`,

		TablePendingWinners,
	)

	_, err := timescaleClient.Exec(
		statementText,
		address,
		network_,
		token.TokenShortName,
		startBlock,
		endBlock,
	)

	if err != nil {
		log.Fatal(func(k *log.Log) {
			k.Context = Context

			k.Format(
				"Failed to mark winnings with token %s for address %s as rewarded!",
				token.TokenShortName,
				address,
			)

			k.Payload = err
		})
	}
}
