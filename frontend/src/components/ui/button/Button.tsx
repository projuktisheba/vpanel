import { ReactNode } from "react";

interface ButtonProps {
  children: ReactNode;
  size?: "xs" | "sm" | "md";
  variant?: "primary" | "outline" | "success" | "danger" | "warning";
  startIcon?: ReactNode;
  endIcon?: ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  className?: string;
  isHidden?: boolean;
}

const Button: React.FC<ButtonProps> = ({
  children,
  size = "md",
  variant = "primary",
  startIcon,
  endIcon,
  onClick,
  className = "",
  disabled = false,
  isHidden = false,
}) => {
  // Size Classes with different border-radius
  const sizeClasses = {
    xs: "px-2 py-1 text-xs rounded-sm",
    sm: "px-4 py-2.5 text-sm rounded-md",
    md: "px-5 py-3.5 text-sm rounded-lg",
  };

  // Variant Classes with dark mode support
  const variantClasses = {
    primary:
      "bg-brand-500 text-white shadow-theme-xs hover:bg-brand-600 disabled:bg-brand-300 dark:bg-brand-600 dark:hover:bg-brand-700 dark:disabled:bg-brand-400",
    outline:
      "bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700 dark:hover:bg-white/[0.03] dark:hover:text-gray-300",
    success:
      "bg-green-500 text-white hover:bg-green-600 disabled:bg-green-300 dark:bg-green-600 dark:hover:bg-green-700 dark:disabled:bg-green-400",
    danger:
      "bg-red-500 text-white hover:bg-red-600 disabled:bg-red-300 dark:bg-red-600 dark:hover:bg-red-700 dark:disabled:bg-red-400",
    warning:
      "bg-yellow-600 text-white shadow-theme-xs hover:bg-yellow-600 disabled:bg-yellow-300 dark:bg-yellow-600 dark:hover:bg-yellow-700 dark:disabled:bg-yellow-400",
  };

  return (
    <button
      className={`inline-flex items-center justify-center gap-2 transition ${className} ${
        sizeClasses[size]
      } ${variantClasses[variant]}  ${isHidden ? "hidden" : "block"}  ${
        disabled ? "cursor-not-allowed opacity-50" : ""
      }`}
      onClick={onClick}
      disabled={disabled}
    >
      {startIcon && <span className="flex items-center">{startIcon}</span>}
      {children}
      {endIcon && <span className="flex items-center">{endIcon}</span>}
    </button>
  );
};

export default Button;
