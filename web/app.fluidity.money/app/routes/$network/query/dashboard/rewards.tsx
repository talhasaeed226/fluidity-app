import type { Chain } from "~/util/chainUtils/chains";
import type { LoaderFunction } from "@remix-run/node";
import type { Rewarders } from "~/util/rewardAggregates";

import { JsonRpcProvider } from "@ethersproject/providers";
import { getTotalPrizePool } from "~/util/chainUtils/ethereum/transaction";
import { json } from "@remix-run/node";
import useApplicationRewardStatistics from "~/queries/useApplicationRewardStatistics";
import { aggregateRewards } from "~/util/rewardAggregates";
import RewardAbi from "~/util/chainUtils/ethereum/RewardPool.json";
import config from "~/webapp.config.server";
import {
  TimeSepUserYield,
  useUserYieldAll,
  useUserYieldByAddress,
} from "~/queries/useUserYield";

export type RewardsLoaderData = {
  network: Chain;
  rewarders: Rewarders;
  rewards: TimeSepUserYield;
  fluidTokenMap: { [symbol: string]: string };
  fluidPairs: number;
  totalPrizePool: number;
  networkFee: number;
  gasFee: number;
  timestamp: number;
};

export const loader: LoaderFunction = async ({ request, params }) => {
  const network = params.network ?? "";
  const fluidPairs = config.config[network ?? ""].fluidAssets.length;

  const networkFee = 0.002;
  const gasFee = 0.002;

  const url = new URL(request.url);
  const address = url.searchParams.get("address");

  try {
    const mainnetId = 0;
    const infuraRpc = config.drivers["ethereum"][mainnetId].rpc.http;

    const provider = new JsonRpcProvider(infuraRpc);

    const rewardPoolAddr = "0xD3E24D732748288ad7e016f93B1dc4F909Af1ba0";

    const { tokens } = config.config[network];

    const fluidTokenMap = tokens.reduce(
      (map, token) =>
        token.isFluidOf
          ? {
              ...map,
              [token.symbol]: token.address,
              [token.symbol.slice(1)]: token.address,
            }
          : map,
      {}
    );

    const [
      totalPrizePool,
      { data: rewardsData, errors: rewardsErr },
      { data: appRewardData, errors: appRewardErrors },
    ] = await Promise.all([
      getTotalPrizePool(provider, rewardPoolAddr, RewardAbi),
      address
        ? useUserYieldByAddress(network, address)
        : useUserYieldAll(network),
      useApplicationRewardStatistics(network ?? ""),
    ]);

    if (rewardsErr || !rewardsData) {
      throw rewardsErr;
    }

    if (appRewardErrors || !appRewardData) {
      throw appRewardErrors;
    }

    const rewarders = aggregateRewards(appRewardData);

    return json({
      network,
      rewarders,
      rewards: rewardsData,
      fluidTokenMap,
      fluidPairs,
      totalPrizePool,
      networkFee,
      gasFee,
      timestamp: new Date().getTime(),
    } as RewardsLoaderData);
  } catch (err) {
    console.log(err);
    throw new Error(`Could not fetch Rewards on ${network}: ${err}`);
  } // Fail silently - for now.
};
