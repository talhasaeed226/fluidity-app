// Copyright 2022 Fluidity Money. All rights reserved. Use of this
// source code is governed by a GPL-style license that can be found in the
// LICENSE.md file.

package main

import (
	database "github.com/fluidity-money/fluidity-app/lib/databases/timescale/tvl"
	queue "github.com/fluidity-money/fluidity-app/lib/queues/ido"
)

func main() {
	queue.TvlUpdates(database.InsertTvl)
}
