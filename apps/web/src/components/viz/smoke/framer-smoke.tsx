'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { Button } from '@carbon/react';
import { useVizTheme } from '../theme';
import styles from './smoke.module.scss';

export function FramerSmoke() {
  const [active, setActive] = useState(false);
  const tokens = useVizTheme();

  return (
    <div className={styles.framerStage}>
      <motion.div
        className={styles.framerPulse}
        style={{ background: active ? tokens['support-error'] : tokens.interactive }}
        animate={{ scale: active ? [1, 1.4, 1] : 1, opacity: active ? [1, 0.6, 1] : 1 }}
        transition={{ duration: 0.6, repeat: active ? Infinity : 0, ease: 'easeInOut' }}
      />
      <Button kind="tertiary" size="sm" onClick={() => setActive((v) => !v)}>
        {active ? 'Stop propagation' : 'Propagate failure'}
      </Button>
    </div>
  );
}
