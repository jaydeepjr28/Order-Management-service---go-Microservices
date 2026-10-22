import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { toast } from 'sonner';
import { GoogleIcon } from '@/components/ui/GoogleIcon';
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react';

export function OAuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const setOAuthTokens = useAuthStore((state) => state.setOAuthTokens);
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    const error = searchParams.get('error');
    if (error) {
      setStatus('error');
      setErrorMessage(error);
      toast.error('Google Sign-In failed', {
        description: error,
      });
      const timer = setTimeout(() => navigate('/login'), 2500);
      return () => clearTimeout(timer);
    }

    const accessToken = searchParams.get('access_token');
    const refreshToken = searchParams.get('refresh_token');
    const customerID = searchParams.get('customer_id');
    const role = (searchParams.get('role') as 'customer' | 'admin') || 'customer';
    const name = searchParams.get('name') || '';
    const redirectUrl = searchParams.get('redirect') || '/';

    if (!accessToken || !customerID) {
      setStatus('error');
      setErrorMessage('Missing authentication credentials in Google callback.');
      toast.error('Authentication Error', {
        description: 'No session tokens received. Please try again.',
      });
      const timer = setTimeout(() => navigate('/login'), 2500);
      return () => clearTimeout(timer);
    }

    // Persist session tokens to state and local storage
    try {
      setOAuthTokens({
        accessToken,
        refreshToken: refreshToken || '',
        customerID,
      });
      setStatus('success');
      toast.success(name ? `Welcome, ${name}!` : 'Signed in with Google successfully!');

      const destination = role === 'admin' && redirectUrl === '/' ? '/admin' : redirectUrl;
      const timer = setTimeout(() => {
        navigate(destination, { replace: true });
      }, 900);
      return () => clearTimeout(timer);
    } catch (err: any) {
      setStatus('error');
      setErrorMessage(err.message || 'Failed to initialize session');
      const timer = setTimeout(() => navigate('/login'), 2500);
      return () => clearTimeout(timer);
    }
  }, [searchParams, navigate, setOAuthTokens]);

  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md p-8 rounded-2xl border border-border bg-card shadow-elevated text-center space-y-6">
        <div className="w-16 h-16 rounded-2xl bg-surface border border-border/80 flex items-center justify-center mx-auto shadow-sm">
          <GoogleIcon className="w-8 h-8" />
        </div>

        <div className="space-y-2">
          <h1 className="text-xl font-bold tracking-tight text-foreground">
            {status === 'loading' && 'Authenticating with Google...'}
            {status === 'success' && 'Signed in successfully!'}
            {status === 'error' && 'Authentication Failed'}
          </h1>
          <p className="text-xs text-muted-foreground">
            {status === 'loading' && 'Securely syncing your STRATUM account and session credentials.'}
            {status === 'success' && 'Redirecting you to your destination...'}
            {status === 'error' && (errorMessage || 'Returning to sign-in page...')}
          </p>
        </div>

        <div className="flex items-center justify-center py-2">
          {status === 'loading' && (
            <div className="flex items-center gap-2.5 text-xs font-medium text-primary">
              <Loader2 className="w-5 h-5 animate-spin" />
              <span>Verifying authorization payload...</span>
            </div>
          )}

          {status === 'success' && (
            <div className="flex items-center gap-2 text-xs font-semibold text-emerald-500">
              <CheckCircle2 className="w-5 h-5" />
              <span>Session verified!</span>
            </div>
          )}

          {status === 'error' && (
            <div className="flex items-center gap-2 text-xs font-semibold text-red-500">
              <AlertCircle className="w-5 h-5" />
              <span>Redirecting to login...</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
