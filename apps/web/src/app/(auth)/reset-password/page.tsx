'use client';

import { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Button, InlineNotification, PasswordInput } from '@carbon/react';
import { resetPassword } from '@/lib/auth';

export default function ResetPasswordPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';

  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (password !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    if (!token) {
      setError('Invalid or missing reset token.');
      return;
    }
    setError('');
    setLoading(true);
    try {
      await resetPassword(token, password);
      setSuccess(true);
      setTimeout(() => router.push('/login'), 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Reset failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="login-page">
      <div className="login-container">
        <div className="login-brand">
          <span className="login-brand__icon">⚡</span>
          <h1 className="login-brand__name">ChaosPlane</h1>
          <p className="login-brand__tagline">Chaos engineering, simplified.</p>
        </div>

        <div className="login-card">
          <h2 className="login-card__title">Set new password</h2>

          {error && (
            <InlineNotification
              kind="error"
              title="Error"
              subtitle={error}
              lowContrast
              hideCloseButton={false}
              onCloseButtonClick={() => setError('')}
            />
          )}

          {success ? (
            <InlineNotification
              kind="success"
              title="Password updated"
              subtitle="Your password has been reset. Redirecting to sign in..."
              lowContrast
              hideCloseButton
            />
          ) : (
            <form className="login-form" onSubmit={handleSubmit}>
              <PasswordInput
                id="password"
                labelText="New password"
                placeholder="••••••••"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
              <PasswordInput
                id="confirm"
                labelText="Confirm new password"
                placeholder="••••••••"
                autoComplete="new-password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                required
              />
              <Button type="submit" className="login-form__submit" disabled={loading || !token}>
                {loading ? 'Updating...' : 'Update password'}
              </Button>
            </form>
          )}

          <p className="login-footer">
            <a href="/login" className="login-link">Back to sign in</a>
          </p>
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

        .login-form {
          display: flex;
          flex-direction: column;
          gap: var(--cds-spacing-05);
        }

        .login-form__submit {
          width: 100%;
          max-width: 100%;
          justify-content: center;
          margin-top: var(--cds-spacing-02);
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
