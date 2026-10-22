import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { Compass } from 'lucide-react';

export function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <div className="min-h-[70vh] flex items-center justify-center p-6 text-center">
      <div className="max-w-md space-y-5">
        <div className="w-16 h-16 rounded-2xl bg-primary-light text-primary flex items-center justify-center mx-auto">
          <Compass className="w-8 h-8" />
        </div>
        <h1 className="text-4xl font-bold font-editorial text-foreground">
          Page Not Found
        </h1>
        <p className="text-sm text-muted-foreground leading-relaxed">
          The requested page could not be located on the platform. It might have been moved or removed.
        </p>
        <div className="pt-2 flex justify-center gap-3">
          <Button onClick={() => navigate('/')} variant="primary" size="md">
            Return Home
          </Button>
          <Button onClick={() => navigate('/products')} variant="outline" size="md">
            Browse Menu
          </Button>
        </div>
      </div>
    </div>
  );
}
