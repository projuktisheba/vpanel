interface PreloaderProps {
  message?: string;
  size?: number;          // spinner size (default 48px)
  color?: string;         // border-top color (Tailwind class)
  overlay?: boolean;      // add overlay background
  className?: string;     // extra classes if needed
}

export function Preloader({
  message = "",
  size = 48,
  color = "border-t-blue-600",
  overlay = true,
  className = "",
}: PreloaderProps) {
  return (
    <div
      className={`${
        overlay ? "fixed inset-0 bg-white/70 dark:bg-gray-900/70" : ""
      } z-50 flex items-center justify-center ${className}`}
    >
      <div className="text-center">
        <div
          className={`animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto dark:border-gray-400 ${color} mx-auto`}
          style={{
            height: size,
            width: size,
          }}
        ></div>

        {message && (
          <p className="mt-4 text-gray-700 dark:text-gray-300 text-sm">
            {message}
          </p>
        )}
      </div>
    </div>
  );
}
