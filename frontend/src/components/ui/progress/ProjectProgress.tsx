import { Loader } from "lucide-react";
import React from "react";

export interface Step {
  title: string;
  description?: string;
  hasError: boolean;
}

interface ProjectProgressProps {
  steps: Step[];
  currentStep: number;
}

const ProjectProgress: React.FC<ProjectProgressProps> = ({
  steps,
  currentStep,
}) => {
  const globalError = steps
    .slice(0, currentStep + 1)
    .some((step) => step.hasError);

  return (
    <div className="mt-4 mx-auto max-w-5xl space-y-8">
      <div className="rounded-md border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 shadow-default">
        <div className="p-4 sm:p-6 xl:p-10">
          <div className="relative flex flex-col md:flex-row items-start md:items-center justify-between w-full">

            {/* BG Line */}
            <div className="absolute top-4 left-0 hidden h-1 w-full -translate-y-1/2 bg-gray-200 dark:bg-gray-700 md:block"></div>

            {/* Active Progress Line */}
            <div
              className={`absolute top-4 left-0 hidden h-1 -translate-y-1/2 transition-all duration-500 ease-in-out md:block ${
                globalError ? "bg-red-500" : "bg-indigo-500"
              }`}
              style={{
                width: `${
                  currentStep === 0 ? 0 : (currentStep / steps.length) * 100
                }%`,
              }}
            ></div>

            {steps.map((step: Step, index: number) => {
              const isCompleted = index < currentStep;
              const isCurrent = index === currentStep;
              const isPending = index > currentStep;
              const stepHasError = step.hasError || globalError;

              return (
                <div
                  key={index}
                  className={`relative z-10 flex flex-1 flex-row md:flex-col items-center md:justify-center gap-4 md:gap-0 ${
                    index !== 0 ? "mt-6 md:mt-0" : ""
                  }`}
                >
                  {/* Vertical line for mobile */}
                  {index !== steps.length - 1 && (
                    <div
                      className={`absolute left-4 top-8 h-full w-0.5 -translate-x-1/2 md:hidden ${
                        stepHasError
                          ? "bg-red-500"
                          : isCompleted
                          ? "bg-indigo-500"
                          : "bg-gray-300 dark:bg-gray-700"
                      }`}
                    ></div>
                  )}

                  {/* Circle */}
                  <div
                    className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2 transition-all duration-300
                      ${
                        stepHasError
                          ? "border-red-500 bg-red-100 dark:bg-red-900/40"
                          : ""
                      }
                      ${
                        !stepHasError && isCompleted
                          ? "border-indigo-500 bg-indigo-500"
                          : ""
                      }
                      ${
                        !stepHasError && isCurrent
                          ? "border-indigo-500 bg-white dark:bg-gray-800"
                          : ""
                      }
                      ${
                        isPending || (stepHasError && isCurrent)
                          ? "border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800"
                          : ""
                      }
                    `}
                  >
                    {isCompleted &&
                      (stepHasError ? (
                        <div className="h-2.5 w-2.5 rounded-full bg-red-500" />
                      ) : (
                        <div className="h-2.5 w-2.5 rounded-full bg-white" />
                      ))}

                    {isCurrent &&
                      (stepHasError ? (
                        <div className="h-2.5 w-2.5 rounded-full bg-gray-300 dark:bg-gray-600" />
                      ) : (
                        <Loader className="h-4 w-4 animate-spin text-indigo-600 dark:text-indigo-400" />
                      ))}

                    {isPending && (
                      <div className="h-2.5 w-2.5 rounded-full bg-gray-300 dark:bg-gray-600" />
                    )}
                  </div>

                  {/* Text */}
                  <div className="flex flex-col md:items-center md:pt-4 md:text-center">
                    <span
                      className={`text-sm font-bold transition-colors duration-300 ${
                        isCompleted
                          ? stepHasError
                            ? "text-red-500"
                            : "text-gray-900 dark:text-gray-200"
                          : "text-gray-500 dark:text-gray-400"
                      }`}
                    >
                      {step.title}
                    </span>
                    <span className="text-xs font-medium text-gray-500 dark:text-gray-500 mt-0.5">
                      {step.description}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProjectProgress;
