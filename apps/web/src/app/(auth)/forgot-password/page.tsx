'use client';

import { useState } from 'react';
import { Button, InlineNotification, TextInput } from '@carbon/react';
import { forgotPassword } from '@/lib/auth';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await forgotPassword(email);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Request failed');
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
          <h2 className="login-card__title">Reset password</h2>
          <p className="login-card__subtitle">
            Enter your email and we&apos;ll send you a link to reset your password.
          </p>

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
              title="Check your email"
              subtitle="If an account exists for that address, we sent a password reset link."
              lowContrast
              hideCloseButton
            />
          ) : (
            <form className="login-form" onSubmit={handleSubmit}>
              <TextInput
                id="email"
                labelText="Email address"
                type="email"
                placeholder="you@company.com"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
              <Button type="submit" className="login-form__submit" disabled={loading}>
                {loading ? 'Sending...' : 'Send reset link'}
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

        .login-card__subtitle {
          color: var(--cds-text-secondary);
          font-size: var(--cds-body-short-01-font-size);
          margin: calc(-1 * var(--cds-spacing-04)) 0 0;
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
