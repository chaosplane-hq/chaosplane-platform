'use client';

import { Suspense, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Button, InlineNotification, Loading, PasswordInput, TextInput } from '@carbon/react';
import { lookupInvitation, acceptInvitation, type InvitationDetails } from '@/lib/auth';

type LoadState = 'loading' | 'ready' | 'error';

export default function AcceptInvitationPage() {
  return <Suspense fallback={<Loading withOverlay />}><AcceptInvitationContent /></Suspense>;
}

function AcceptInvitationContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';

  const [loadState, setLoadState] = useState<LoadState>('loading');
  const [invitation, setInvitation] = useState<InvitationDetails | null>(null);
  const [loadError, setLoadError] = useState('');

  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!token) {
      setLoadError('No invitation token found in the URL.');
      setLoadState('error');
      return;
    }
    lookupInvitation(token)
      .then((data) => {
        setInvitation(data);
        setLoadState('ready');
      })
      .catch((err: unknown) => {
        setLoadError(err instanceof Error ? err.message : 'Failed to load invitation');
        setLoadState('error');
      });
  }, [token]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await acceptInvitation(token, name, password);
      router.push('/experiments');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to accept invitation');
    } finally {
      setSubmitting(false);
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
          <h2 className="login-card__title">Accept invitation</h2>

          {loadState === 'loading' && (
            <div className="verify-loading">
              <Loading small withOverlay={false} description="Loading invitation..." />
              <p className="verify-loading__text">Loading invitation details...</p>
            </div>
          )}

          {loadState === 'error' && (
            <>
              <InlineNotification
                kind="error"
                title="Invalid invitation"
                subtitle={loadError}
                lowContrast
                hideCloseButton
              />
              <p className="login-footer">
                <a href="/login" className="login-link">Back to sign in</a>
              </p>
            </>
          )}

          {loadState === 'ready' && invitation && (
            <>
              <div className="invite-details">
                <p className="invite-details__text">
                  <strong>{invitation.inviterName}</strong> invited you to join{' '}
                  <strong>{invitation.tenantName}</strong> as <span className="invite-details__role">{invitation.role}</span>.
                </p>
                <p className="invite-details__email">Joining as: {invitation.email}</p>
              </div>

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

              <form className="login-form" onSubmit={handleSubmit}>
                <TextInput
                  id="name"
                  labelText="Full name"
                  placeholder="Jane Smith"
                  autoComplete="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                />
                <PasswordInput
                  id="password"
                  labelText="Create a password"
                  placeholder="••••••••"
                  autoComplete="new-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
                <Button type="submit" className="login-form__submit" disabled={submitting}>
                  {submitting ? 'Joining...' : 'Accept & join team'}
                </Button>
              </form>

              <p className="login-footer">
                Already have an account?{' '}
                <a href="/login" className="login-link">Sign in</a>
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

        .invite-details {
          background-color: var(--cds-layer-02);
          padding: var(--cds-spacing-05);
          border-left: 3px solid var(--cds-interactive);
        }

        .invite-details__text {
          color: var(--cds-text-primary);
          font-size: var(--cds-body-short-01-font-size);
          margin: 0 0 var(--cds-spacing-02);
        }

        .invite-details__role {
          color: var(--cds-link-primary);
          text-transform: capitalize;
        }

        .invite-details__email {
          color: var(--cds-text-secondary);
          font-size: var(--cds-label-01-font-size);
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
