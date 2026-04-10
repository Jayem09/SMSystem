interface SkeletonProps {
  variant?: 'text' | 'card' | 'table' | 'product';
  count?: number;
  className?: string;
}

const skeletonBase = 'animate-pulse bg-gray-200 rounded';

export function Skeleton({ variant = 'text', count = 1, className = '' }: SkeletonProps) {
  const variants = {
    text: 'h-4 w-full',
    card: 'h-32 w-full rounded-lg',
    table: 'h-12 w-full',
    product: 'h-24 w-full rounded-lg',
  };

  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className={`${skeletonBase} ${variants[variant]} ${className}`}
        />
      ))}
    </>
  );
}

export function SkeletonCard({ className = '' }: { className?: string }) {
  return (
    <div className={`bg-white rounded-lg border border-gray-200 p-4 ${className}`}>
      <Skeleton variant="text" className="w-3/4 mb-2" />
      <Skeleton variant="text" className="w-1/2 mb-4" />
      <Skeleton variant="text" />
    </div>
  );
}

export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-2">
      <Skeleton variant="table" />
      {Array.from({ length: rows }).map((_, i) => (
        <Skeleton key={i} variant="table" />
      ))}
    </div>
  );
}
