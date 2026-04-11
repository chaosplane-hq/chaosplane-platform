'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Button, InlineNotification, Loading } from '@carbon/react';
import { verifyEmail } from '@/lib/auth';

type State = 'verifying' | 'success' | 'expired' | 'error';

export default function VerifyEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';

  const [state, setState] = useState<State>('verifying');
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (!token) {
      setState('error');
      setErrorMsg('No verification token found in the URL.');
      return;
    }

    verifyEmail(token)
      .then(() => {
        setState('success');
        setTimeout(() => router.push('/experiments'), 2500);
      })
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : 'Verification failed';
        if (msg.toLowerCase().includes('expir')) {
          setState('expired');
        } else {
          setState('error');
          setErrorMsg(msg);
        }
      });
  }, [token, router]);

  return (
    <main className="login-page">
      <div className="login-container">
        <div className="login-brand">
          <span className="login-brand__icon">⚡</span>
          <h1 className="login-brand__name">ChaosPlane</h1>
          <p className="login-brand__tagline">Chaos engineering, simplified.</p>
        </div>

        <div className="login-card">
          <h2 className="login-card__title">Email verification</h2>

          {state === 'verifying' && (
            <div className="verify-loading">
              <Loading small withOverlay={false} description="Verifying your email..." />
              <p className="verify-loading__text">Verifying your email address...</p>
            </div>
          )}

          {state === 'success' && (
            <InlineNotification
              kind="success"
              title="Email verified"
              subtitle="Your email has been verified. Redirecting to dashboard..."
              lowContrast
              hideCloseButton
            />
          )}

          {state === 'expired' && (
            <>
              <InlineNotification
                kind="warning"
                title="Link expired"
                subtitle="This verification link has expired. Request a new one from your account settings."
                lowContrast
                hideCloseButton
              />
              <Button kind="tertiary" className="login-form__submit" onClick={() => router.push('/login')}>
                Back to sign in
              </Button>
            </>
          )}

          {state === 'error' && (
            <>
              <InlineNotification
                kind="error"
                title="Verification failed"
                subtitle={errorMsg || 'Something went wrong. Please try again.'}
                lowContrast
                hideCloseButton
              />
              <Button kind="tertiary" className="login-form__submit" onClick={() => router.push('/login')}>
                Back to sign in
              </Button>
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

        .login-form__submit {
          width: 100%;
          max-width: 100%;
          justify-content: center;
        }
      `}</style>
    </main>
  );
}
