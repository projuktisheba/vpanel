import { ReactNode } from "react";
import { Modal } from ".";
import Button from "../button/Button";

interface AlertModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  message: string | ReactNode;
  type?: "success" | "error" | "warning";
  primaryAction?: {
    label: string;
    onClick: () => void;
    variant?: "primary" | "outline";
    className?: string;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
    variant?: "primary" | "outline";
    className?: string;
  };
}

export default function AlertModal({
  isOpen,
  onClose,
  title,
  message,
  type = "warning",
  primaryAction,
  secondaryAction,
}: AlertModalProps) {
  // Set colors based on type
  const typeColors = {
    success: {
      bg: "bg-green-50 dark:bg-green-900/20",
      border: "border-green-300 dark:border-green-700",
      text: "text-green-700 dark:text-green-400",
      textMessage: "text-green-600 dark:text-green-300",
    },
    error: {
      bg: "bg-red-50 dark:bg-red-900/20",
      border: "border-red-300 dark:border-red-700",
      text: "text-red-700 dark:text-red-400",
      textMessage: "text-red-600 dark:text-red-300",
    },
    warning: {
      bg: "bg-yellow-50 dark:bg-yellow-900/20",
      border: "border-yellow-300 dark:border-yellow-700",
      text: "text-yellow-700 dark:text-yellow-400",
      textMessage: "text-yellow-600 dark:text-yellow-300",
      icon: "⚠️",
    },
  };

  const colors = typeColors[type];

  return (
    <Modal isOpen={isOpen} onClose={onClose} className="max-w-[500px] m-4">
      <div className="relative w-full rounded-3xl bg-white p-6 dark:bg-gray-900 lg:p-8">
        {/* Header */}
        {title && (
          <div className="px-2">
            <h4 className="mb-3 text-2xl font-semibold text-gray-800 dark:text-white/90">
              {title}
            </h4>
          </div>
        )}

        {/* Message Section */}
        <div className={`mt-4 px-2 rounded-xl border ${colors.border} ${colors.bg} p-4`}>
          <h5 className={`mb-2 text-lg font-medium ${colors.text}`}>
            {type.charAt(0).toUpperCase() + type.slice(1)}
          </h5>
          <p className={`text-sm ${colors.textMessage}`}>{message}</p>
        </div>

        {/* Footer Buttons */}
        <div className="mt-6 flex items-center gap-3 px-2 lg:justify-end">
          {secondaryAction && (
            <Button
              size="sm"
              variant={secondaryAction.variant || "outline"}
              className={secondaryAction.className}
              onClick={secondaryAction.onClick}
            >
              {secondaryAction.label}
            </Button>
          )}
          {primaryAction && (
            <Button
              size="sm"
              variant={primaryAction.variant || "primary"}
              className={primaryAction.className}
              onClick={primaryAction.onClick}
            >
              {primaryAction.label}
            </Button>
          )}
        </div>
      </div>
    </Modal>
  );
}
