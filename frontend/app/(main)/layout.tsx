import { Navigation } from '@/components/Navigation';

export default function MainLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="flex min-h-screen">
      {/* Main Content */}
      <main className="flex-1">
        <Navigation />
        {children}
      </main>
    </div>
  );
}
