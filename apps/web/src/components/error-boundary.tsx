'use client';

import { Component, type ReactNode } from 'react';
import { Button, Tile } from '@carbon/react';
import { Renew } from '@carbon/icons-react';
import { t } from '@/lib/i18n';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  handleReset = () => {
    this.setState({ hasError: false });
  };

  render() {
    if (this.state.hasError) {
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
              ⚠
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
            <Button onClick={this.handleReset} renderIcon={Renew} kind="primary">
              {t('error.500.action')}
            </Button>
          </Tile>
        </main>
      );
    }

    return this.props.children;
  }
}
