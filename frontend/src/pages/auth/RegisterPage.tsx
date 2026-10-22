import React, { useState, useEffect, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/useAuthStore';
import { authApi } from '@/lib/api/services';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Mail, Lock, User, KeyRound, ArrowLeft, RefreshCw, CheckCircle2 } from 'lucide-react';
import { GoogleIcon } from '@/components/ui/GoogleIcon';

const registerSchema = z
  .object({
    name: z.string().min(2, 'Name must be at least 2 characters'),
    email: z.string().email('Please enter a valid email address'),
    password: z.string().min(6, 'Password must be at least 6 characters'),
    confirmPassword: z.string().min(6, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  });

type RegisterFormValues = z.infer<typeof registerSchema>;

export function RegisterPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const redirectUrl = searchParams.get('redirect') || '/';
  const { verifyOTPAndRegister } = useAuthStore();

  // Step state: 'details' -> 'otp'
  const [step, setStep] = useState<'details' | 'otp'>('details');
  const [submittedEmail, setSubmittedEmail] = useState('');
  const [otp, setOtp] = useState(['', '', '', '', '', '']);
  const [cooldown, setCooldown] = useState(60);
  const [apiError, setApiError] = useState<string | null>(null);
  const [isVerifying, setIsVerifying] = useState(false);
  const [isResending, setIsResending] = useState(false);

  const otpInputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      name: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
  });

  // Countdown timer for resending OTP
  useEffect(() => {
    if (step === 'otp' && cooldown > 0) {
      const timer = setInterval(() => setCooldown((prev) => prev - 1), 1000);
      return () => clearInterval(timer);
    }
  }, [step, cooldown]);

  // Handle Step 1: Send OTP
  const onSendOTP = async (values: RegisterFormValues) => {
    setApiError(null);
    try {
      await authApi.sendRegistrationOTP({
        name: values.name,
        email: values.email,
        password: values.password,
      });

      setSubmittedEmail(values.email);
      setCooldown(60);
      setStep('otp');
      setTimeout(() => otpInputRefs.current[0]?.focus(), 100);
    } catch (err: any) {
      setApiError(err.message || 'Failed to send verification code. Email may already be in use.');
    }
  };

  // Handle 6-digit OTP input typing & auto-focus
  const handleOtpChange = (index: number, val: string) => {
    const clean = val.replace(/\D/g, '');
    if (!clean) {
      const updated = [...otp];
      updated[index] = '';
      setOtp(updated);
      return;
    }

    // Support pasting full code
    if (clean.length > 1) {
      const chars = clean.slice(0, 6).split('');
      const updated = [...otp];
      chars.forEach((c, i) => {
        if (index + i < 6) updated[index + i] = c;
      });
      setOtp(updated);
      const nextFocus = Math.min(index + clean.length, 5);
      otpInputRefs.current[nextFocus]?.focus();
      return;
    }

    const updated = [...otp];
    updated[index] = clean;
    setOtp(updated);

    if (index < 5 && clean) {
      otpInputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !otp[index] && index > 0) {
      otpInputRefs.current[index - 1]?.focus();
    }
  };

  // Handle Step 2: Verify OTP & Activate
  const onVerifyOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    const fullOtp = otp.join('');
    if (fullOtp.length !== 6) {
      setApiError('Please enter all 6 digits of the verification code.');
      return;
    }

    setIsVerifying(true);
    setApiError(null);
    try {
      await verifyOTPAndRegister({
        email: submittedEmail,
        otp: fullOtp,
      });

      const role = useAuthStore.getState().role;
      if (role === 'admin' && redirectUrl === '/') {
        navigate('/admin');
      } else {
        navigate(redirectUrl);
      }
    } catch (err: any) {
      setApiError(err.message || 'Invalid or expired verification code.');
    } finally {
      setIsVerifying(false);
    }
  };

  const handleResend = async () => {
    if (cooldown > 0 || isResending) return;
    setIsResending(true);
    setApiError(null);
    try {
      await authApi.resendRegistrationOTP(submittedEmail);
      setCooldown(60);
      setOtp(['', '', '', '', '', '']);
      otpInputRefs.current[0]?.focus();
    } catch (err: any) {
      setApiError(err.message || 'Failed to resend code.');
    } finally {
      setIsResending(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center space-y-2">
          <Link to="/" className="inline-flex items-center gap-2 font-bold text-xl text-foreground">
            <span className="w-8 h-8 rounded-lg bg-primary text-primary-foreground flex items-center justify-center font-bold text-sm">
              S
            </span>
            <span>STRATUM</span>
          </Link>
          <h1 className="text-2xl font-bold font-editorial tracking-tight text-foreground">
            {step === 'details' ? 'Create an account' : 'Verify your email'}
          </h1>
          <p className="text-xs text-muted-foreground">
            {step === 'details'
              ? 'Sign up to build your cart, checkout with verified stock, and track orders.'
              : `We sent a 6-digit confirmation code to ${submittedEmail}`}
          </p>
        </div>

        <Card className="p-6 sm:p-8 border border-border shadow-elevated">
          {apiError && (
            <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-600 dark:text-red-400 text-xs font-medium">
              {apiError}
            </div>
          )}

          {step === 'details' ? (
            <form onSubmit={handleSubmit(onSendOTP)} className="space-y-4">
              <Input
                label="Full Name"
                type="text"
                placeholder="Alex Mercer"
                leftIcon={<User className="w-4 h-4" />}
                error={errors.name?.message}
                {...register('name')}
              />

              <Input
                label="Email Address"
                type="email"
                placeholder="alex@example.com"
                leftIcon={<Mail className="w-4 h-4" />}
                error={errors.email?.message}
                {...register('email')}
              />

              <Input
                label="Password"
                type="password"
                placeholder="••••••••"
                leftIcon={<Lock className="w-4 h-4" />}
                error={errors.password?.message}
                {...register('password')}
              />

              <Input
                label="Confirm Password"
                type="password"
                placeholder="••••••••"
                leftIcon={<Lock className="w-4 h-4" />}
                error={errors.confirmPassword?.message}
                {...register('confirmPassword')}
              />

              <Button
                type="submit"
                isLoading={isSubmitting}
                variant="primary"
                size="lg"
                className="w-full mt-2"
              >
                Send Verification Code
              </Button>

              <div className="relative my-4 pt-1">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-border" />
                </div>
                <div className="relative flex justify-center text-[10px] uppercase tracking-wider">
                  <span className="bg-card px-2.5 text-muted-foreground font-semibold">Or sign up with</span>
                </div>
              </div>

              <Button
                type="button"
                variant="outline"
                size="lg"
                className="w-full font-medium border-border/80 hover:bg-surface/80 transition-all shadow-subtle hover:border-primary/40"
                onClick={() => {
                  window.location.href = `/api/auth/google/login?redirect=${encodeURIComponent(redirectUrl)}`;
                }}
                leftIcon={<GoogleIcon className="w-4 h-4" />}
              >
                Continue with Google
              </Button>
            </form>
          ) : (
            <form onSubmit={onVerifyOTP} className="space-y-6">
              <div className="text-center space-y-2">
                <div className="w-12 h-12 rounded-full bg-primary/10 text-primary mx-auto flex items-center justify-center">
                  <KeyRound className="w-6 h-6" />
                </div>
                <p className="text-xs text-muted-foreground">
                  Enter the 6-digit code dispatched to your inbox.
                </p>
              </div>

              {/* 6-box PIN input */}
              <div className="flex justify-center gap-2 sm:gap-2.5">
                {otp.map((digit, idx) => (
                  <input
                    key={idx}
                    ref={(el) => {
                      otpInputRefs.current[idx] = el;
                    }}
                    type="text"
                    inputMode="numeric"
                    maxLength={idx === 0 ? 6 : 1}
                    value={digit}
                    onChange={(e) => handleOtpChange(idx, e.target.value)}
                    onKeyDown={(e) => handleKeyDown(idx, e)}
                    className="w-11 h-13 sm:w-12 sm:h-14 text-center text-xl font-mono font-bold rounded-lg border border-border bg-surface text-foreground focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all shadow-sm"
                  />
                ))}
              </div>

              <div className="space-y-3 pt-2">
                <Button
                  type="submit"
                  isLoading={isVerifying}
                  variant="primary"
                  size="lg"
                  className="w-full"
                  leftIcon={<CheckCircle2 className="w-4 h-4" />}
                >
                  Verify & Create Account
                </Button>

                <div className="flex items-center justify-between text-xs pt-1">
                  <button
                    type="button"
                    onClick={() => {
                      setStep('details');
                      setApiError(null);
                    }}
                    className="inline-flex items-center gap-1 text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <ArrowLeft className="w-3.5 h-3.5" />
                    <span>Change details</span>
                  </button>

                  {cooldown > 0 ? (
                    <span className="text-muted-foreground font-mono">
                      Resend in {cooldown}s
                    </span>
                  ) : (
                    <button
                      type="button"
                      onClick={handleResend}
                      disabled={isResending}
                      className="inline-flex items-center gap-1 font-semibold text-primary hover:underline"
                    >
                      <RefreshCw className={`w-3 h-3 ${isResending ? 'animate-spin' : ''}`} />
                      <span>Resend code</span>
                    </button>
                  )}
                </div>
              </div>
            </form>
          )}
        </Card>

        <p className="text-center text-xs text-muted-foreground">
          Already have an account?{' '}
          <Link
            to={`/login${redirectUrl !== '/' ? `?redirect=${encodeURIComponent(redirectUrl)}` : ''}`}
            className="font-semibold text-primary hover:underline"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
