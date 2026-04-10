'use client';

import { Button, Tile } from '@carbon/react';
import { Renew } from '@carbon/icons-react';
import { t } from '@/lib/i18n';

interface GlobalErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function GlobalError({ error, reset }: GlobalErrorProps) {
  return (
    <html lang="en">
      <body style={{ margin: 0, backgroundColor: '#161616' }}>
        <main
          style={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: 'var(--cds-background, #161616)',
            padding: '2rem',
          }}
        >
          <Tile
            style={{
              maxWidth: '480px',
              width: '100%',
              padding: '3rem',
              textAlign: 'center',
            }}
          >
            <p
              style={{
                fontSize: '6rem',
                fontWeight: 700,
                lineHeight: 1,
                margin: '0 0 1.5rem',
                color: 'var(--cds-support-error, #fa4d56)',
                letterSpacing: '-0.04em',
              }}
            >
              ⚠
            </p>
            <h1
              style={{
                fontSize: '1.5rem',
                fontWeight: 600,
                margin: '0 0 1rem',
                color: 'var(--cds-text-primary, #f4f4f4)',
              }}
            >
              {t('error.global.title')}
            </h1>
            <p
              style={{
                margin: '0 0 2rem',
                color: 'var(--cds-text-secondary, #c6c6c6)',
                fontSize: '0.875rem',
              }}
            >
              {t('error.global.description')}
            </p>
            {error.digest && (
              <p
                style={{
                  margin: '0 0 1.5rem',
                  color: 'var(--cds-text-helper, #8d8d8d)',
                  fontSize: '0.75rem',
                }}
              >
                {error.digest}
              </p>
            )}
            <Button onClick={reset} renderIcon={Renew} kind="primary">
              {t('error.global.action')}
            </Button>
          </Tile>
        </main>
      </body>
    </html>
  );
}
