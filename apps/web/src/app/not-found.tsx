import { Button, Tile } from '@carbon/react';
import { ArrowLeft } from '@carbon/icons-react';
import Link from 'next/link';
import { t } from '@/lib/i18n';

export default function NotFound() {
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
            color: 'var(--cds-text-primary)',
            letterSpacing: '-0.04em',
          }}
        >
          404
        </p>
        <h1
          style={{
            fontSize: '1.5rem',
            fontWeight: 600,
            margin: '0 0 var(--cds-spacing-04)',
            color: 'var(--cds-text-primary)',
          }}
        >
          {t('error.404.title')}
        </h1>
        <p
          style={{
            margin: '0 0 var(--cds-spacing-07)',
            color: 'var(--cds-text-secondary)',
            fontSize: '0.875rem',
          }}
        >
          {t('error.404.description')}
        </p>
        <Button as={Link} href="/" renderIcon={ArrowLeft} kind="primary">
          {t('error.404.action')}
        </Button>
      </Tile>
    </main>
  );
}
