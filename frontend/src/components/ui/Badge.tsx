import React from 'react';
import { cn } from '@/lib/utils';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'pending' | 'paid' | 'ready' | 'neutral' | 'danger' | 'primary';
  size?: 'sm' | 'md';
}

export function Badge({
  className,
  variant = 'neutral',
  size = 'md',
  children,
  ...props
}: BadgeProps) {
  const variants = {
    pending:
      'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
    paid:
      'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20',
    ready:
      'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
    neutral:
      'bg-neutral-500/10 text-neutral-700 dark:text-neutral-300 border-neutral-500/20',
    danger:
      'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
    primary:
      'bg-primary-light text-primary border-primary/20 font-semibold',
  };

  const sizes = {
    sm: 'text-[11px] px-2 py-0.5',
    md: 'text-xs px-2.5 py-1',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center font-medium rounded-full border tracking-wide uppercase font-mono select-none',
        variants[variant],
        sizes[size],
        className
      )}
      {...props}
    >
      {children}
    </span>
  );
}
