'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { InlineNotification, Loading } from '@carbon/react';
import { oauthCallback } from '@/lib/auth';

export default function CallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const code = searchParams.get('code') ?? '';
  const state = searchParams.get('state') ?? '';
  const provider = searchParams.get('provider') ?? '';
  const errorParam = searchParams.get('error');

  const [error, setError] = useState(errorParam ?? '');

  useEffect(() => {
    if (errorParam) {
      setError(errorParam);
      return;
    }
    if (!code || !state || !provider) {
      setError('Missing required OAuth parameters.');
      return;
    }

    oauthCallback(provider, code, state)
      .then(() => {
        router.push('/experiments');
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : 'OAuth callback failed');
      });
  }, [code, state, provider, errorParam, router]);

  return (
    <main className="login-page">
      <div className="login-container">
        <div className="login-brand">
          <span className="login-brand__icon">⚡</span>
          <h1 className="login-brand__name">ChaosPlane</h1>
          <p className="login-brand__tagline">Chaos engineering, simplified.</p>
        </div>

        <div className="login-card">
          <h2 className="login-card__title">Signing you in</h2>

          {!error ? (
            <div className="verify-loading">
              <Loading small withOverlay={false} description="Completing sign in..." />
              <p className="verify-loading__text">Completing sign in...</p>
            </div>
          ) : (
            <>
              <InlineNotification
                kind="error"
                title="Authentication failed"
                subtitle={error}
                lowContrast
                hideCloseButton
              />
              <p className="login-footer">
                <a href="/login" className="login-link">Back to sign in</a>
              </p>
            </>
          )}
        </div>
      </div>

      <style jsx>{`
        .login-page {
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          background-color: var(--cds-background);
          padding: var(--cds-spacing-05);
        }

        .login-container {
          width: 100%;
          max-width: 400px;
          display: flex;
          flex-direction: column;
          gap: var(--cds-spacing-07);
        }

        .login-brand {
          text-align: center;
        }

        .login-brand__icon {
          font-size: 2.5rem;
          display: block;
          margin-bottom: var(--cds-spacing-03);
        }

        .login-brand__name {
          font-size: var(--cds-heading-04-font-size);
          font-weight: var(--cds-heading-04-font-weight);
          letter-spacing: var(--cds-heading-04-letter-spacing);
          color: var(--cds-text-primary);
          margin: 0 0 var(--cds-spacing-02);
        }

        .login-brand__tagline {
          color: var(--cds-text-secondary);
          font-size: var(--cds-body-short-01-font-size);
          margin: 0;
        }

        .login-card {
          background-color: var(--cds-layer-01);
          padding: var(--cds-spacing-07);
          display: flex;
          flex-direction: column;
          gap: var(--cds-spacing-06);
        }

        .login-card__title {
          font-size: var(--cds-heading-03-font-size);
          font-weight: var(--cds-heading-03-font-weight);
          color: var(--cds-text-primary);
          margin: 0;
        }

        .verify-loading {
          display: flex;
          align-items: center;
          gap: var(--cds-spacing-04);
        }

        .verify-loading__text {
          color: var(--cds-text-secondary);
          font-size: var(--cds-body-short-01-font-size);
          margin: 0;
        }

        .login-footer {
          color: var(--cds-text-secondary);
          font-size: var(--cds-body-short-01-font-size);
          margin: 0;
          text-align: center;
        }

        .login-link {
          color: var(--cds-link-primary);
          text-decoration: none;
        }

        .login-link:hover {
          text-decoration: underline;
        }
      `}</style>
    </main>
  );
}
