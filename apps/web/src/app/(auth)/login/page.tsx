'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { Button, InlineNotification, PasswordInput, TextInput } from '@carbon/react';
import { LogoGithub } from '@carbon/icons-react';
import { login, oauthAuthorize } from '@/lib/auth';

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      router.push('/experiments');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  async function handleOAuth(provider: string) {
    try {
      const url = await oauthAuthorize(provider);
      window.location.href = url;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'OAuth failed');
    }
  }

  return (
    <main className="login-page">
      <div className="login-container">
        <div className="login-brand">
          <Image
            src="/logo.svg"
            alt="ChaosPlane"
            width={56}
            height={56}
            className="login-brand__logo"
            priority
          />
          <h1 className="login-brand__name">ChaosPlane</h1>
          <p className="login-brand__tagline">Chaos engineering, simplified.</p>
        </div>

        <div className="login-card">
          <h2 className="login-card__title">Sign in</h2>

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

          <div className="login-oauth">
            <Button kind="tertiary" className="login-oauth__btn" onClick={() => handleOAuth('google')}>
              Continue with Google
            </Button>
            <Button kind="tertiary" renderIcon={LogoGithub} iconDescription="GitHub" className="login-oauth__btn" onClick={() => handleOAuth('github')}>
              Continue with GitHub
            </Button>
            <Button kind="tertiary" className="login-oauth__btn" onClick={() => handleOAuth('microsoft')}>
              Continue with Microsoft
            </Button>
          </div>

          <div className="login-divider">
            <span>or sign in with email</span>
          </div>

          <form className="login-form" onSubmit={handleSubmit}>
            <TextInput
              id="email"
              labelText="Email address"
              type="email"
              placeholder="you@company.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            <PasswordInput
              id="password"
              labelText="Password"
              placeholder="••••••••"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Button type="submit" className="login-form__submit" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign in'}
            </Button>
            <p className="login-form__forgot">
              <a href="/forgot-password" className="login-link">Forgot password?</a>
            </p>
          </form>

          <p className="login-footer">
            Don&apos;t have an account?{' '}
            <a href="/register" className="login-link">Create account</a>
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
          max-width: 26rem;
          display: flex;
          flex-direction: column;
          gap: var(--cds-spacing-08);
        }

        .login-brand {
          text-align: center;
          display: flex;
          flex-direction: column;
          align-items: center;
        }

        .login-brand :global(.login-brand__logo) {
          width: 56px;
          height: 56px;
          margin-bottom: var(--cds-spacing-05);
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
          border: 1px solid var(--cds-border-subtle-01);
          border-top: 2px solid var(--cds-border-interactive);
          padding: var(--cds-spacing-08) var(--cds-spacing-07);
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

        .login-oauth {
          display: flex;
          flex-direction: column;
          gap: var(--cds-spacing-03);
        }

        .login-oauth__btn {
          width: 100%;
          max-width: 100%;
          justify-content: center;
        }

        .login-divider {
          display: flex;
          align-items: center;
          gap: var(--cds-spacing-04);
          color: var(--cds-text-secondary);
          font-size: var(--cds-label-01-font-size);
        }

        .login-divider::before,
        .login-divider::after {
          content: '';
          flex: 1;
          height: 1px;
          background-color: var(--cds-border-subtle-01);
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

        .login-form__forgot {
          text-align: center;
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
