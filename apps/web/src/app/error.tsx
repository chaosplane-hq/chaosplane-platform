'use client';

import { Button, Tile } from '@carbon/react';
import { Renew } from '@carbon/icons-react';
import { t } from '@/lib/i18n';

interface ErrorPageProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function ErrorPage({ error, reset }: ErrorPageProps) {
  return (
    <main
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'var(--cds-background)',
        padding: 'var(--cds-spacing-07)',
      }}
    >
      <Tile
        style={{
          maxWidth: '480px',
          width: '100%',
          padding: 'var(--cds-spacing-08)',
          textAlign: 'center',
        }}
      >
        <p
          style={{
            fontSize: '6rem',
            fontWeight: 700,
            lineHeight: 1,
            margin: '0 0 var(--cds-spacing-05)',
            color: 'var(--cds-support-error)',
            letterSpacing: '-0.04em',
          }}
        >
          500
        </p>
        <h1
          style={{
            fontSize: '1.5rem',
            fontWeight: 600,
            margin: '0 0 var(--cds-spacing-04)',
            color: 'var(--cds-text-primary)',
          }}
        >
          {t('error.500.title')}
        </h1>
        <p
          style={{
            margin: '0 0 var(--cds-spacing-07)',
            color: 'var(--cds-text-secondary)',
            fontSize: '0.875rem',
          }}
        >
          {t('error.500.description')}
        </p>
        {error.digest && (
          <p
            style={{
              margin: '0 0 var(--cds-spacing-06)',
              color: 'var(--cds-text-helper)',
              fontSize: '0.75rem',
              fontFamily: 'var(--cds-code-01-font-family)',
            }}
          >
            {error.digest}
          </p>
        )}
        <Button onClick={reset} renderIcon={Renew} kind="primary">
          {t('error.500.action')}
        </Button>
      </Tile>
    </main>
  );
}
