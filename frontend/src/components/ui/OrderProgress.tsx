import React from 'react';
import { cn } from '@/lib/utils';
import { Clock, CreditCard, ChefHat, Check, XCircle } from 'lucide-react';
import { OrderStatus } from '@/types';

interface OrderProgressProps {
  status: OrderStatus;
  className?: string;
}

export function OrderProgress({ status, className }: OrderProgressProps) {
  const normalized = (status || 'pending').toLowerCase();

  if (normalized === 'cancelled' || normalized === 'refunded') {
    return (
      <div className={cn('w-full py-3 text-center', className)}>
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-rose-500/10 border border-rose-500/20 text-rose-500 text-xs font-semibold">
          <XCircle className="w-4 h-4" />
          <span>
            {normalized === 'refunded'
              ? 'Order Cancelled & Refunded (Saga Rollback Completed)'
              : 'Order Cancelled (Compensated via Saga)'}
          </span>
        </div>
      </div>
    );
  }


  const steps = [
    {
      id: 'pending',
      label: 'Order Placed',
      description: 'Pending payment confirmation',
      icon: Clock,
    },
    {
      id: 'paid',
      label: 'Payment Confirmed',
      description: 'Dispatched to kitchen',
      icon: CreditCard,
    },
    {
      id: 'ready',
      label: 'Prepared & Ready',
      description: 'Ready for pickup',
      icon: ChefHat,
    },
  ];

  const getStepStatus = (stepId: string) => {
    if (stepId === 'pending') {
      return 'completed';
    }
    if (stepId === 'paid') {
      if (normalized === 'paid' || normalized === 'ready') return 'completed';
      return 'current';
    }
    if (stepId === 'ready') {
      if (normalized === 'ready') return 'completed';
      return normalized === 'paid' ? 'current' : 'upcoming';
    }
    return 'upcoming';
  };

  return (
    <div className={cn('w-full py-4', className)}>
      <div className="relative flex items-center justify-between">
        {/* Background track line */}
        <div className="absolute top-1/2 left-0 right-0 h-0.5 -translate-y-1/2 bg-border -z-0" />

        {steps.map((step, index) => {
          const stepStatus = getStepStatus(step.id);
          const isCompleted = stepStatus === 'completed';
          const isCurrent = stepStatus === 'current';
          const Icon = step.icon;

          return (
            <div key={step.id} className="relative z-10 flex flex-col items-center">
              <div
                className={cn(
                  'w-10 h-10 rounded-full flex items-center justify-center border-2 transition-all duration-300',
                  isCompleted && 'bg-primary border-primary text-primary-foreground shadow-sm',
                  isCurrent && 'bg-card border-primary text-primary shadow ring-4 ring-primary/10 animate-pulse',
                  !isCompleted && !isCurrent && 'bg-surface border-border text-muted-foreground'
                )}
              >
                {isCompleted && step.id !== normalized ? (
                  <Check className="w-5 h-5 stroke-[2.5]" />
                ) : (
                  <Icon className="w-5 h-5" />
                )}
              </div>
              <div className="mt-3 text-center">
                <span
                  className={cn(
                    'block text-xs font-semibold tracking-tight',
                    isCompleted || isCurrent ? 'text-foreground' : 'text-muted-foreground'
                  )}
                >
                  {step.label}
                </span>
                <span className="hidden sm:block text-[11px] text-muted-foreground mt-0.5">
                  {step.description}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
