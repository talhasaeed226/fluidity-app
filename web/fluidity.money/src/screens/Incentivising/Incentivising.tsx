// Copyright 2022 Fluidity Money. All rights reserved. Use of this
// source code is governed by a GPL-style license that can be found in the
// LICENSE.md file.

import { Display, Heading } from "@fluidity-money/surfing";
import Video from "components/Video";
import useViewport from "hooks/useViewport";
import { isFirefox } from "react-device-detect";
import styles from "./Incentivising.module.scss";

const Incentivising = () => {
  const { width } = useViewport();
  const breakpoint = 860;

  return (
    <div className={styles.container}>
      {width <= breakpoint ? (
      <Video
        src={"/assets/videos/FluidityHowItWorks.mp4"}
        type={"reduce"}
        loop={true}
        className={styles.video}
      />): (
      <Video
        src={"/assets/videos/FluidityHowItWorks.mp4"}
        type={"reduce"}
        loop={true}
        scale={isFirefox ? 1.5 : .6}
        className={styles.video}
      />)}
      <div>
        <div className={styles.blur} />
        <Heading as={"h6"} className={styles.backgroundText}>
          HOW IT WORKS
        </Heading>
        <br />
        <Display
          className={styles.backgroundText}
          size={width > breakpoint ? "lg" : "sm"}
        >
          Incentivising
        </Display>
        <Display
          className={styles.backgroundText}
          size={width > breakpoint ? "lg" : "sm"}
        >
          blockchain utility
        </Display>
      </div>
    </div>
  );
};

export default Incentivising;
