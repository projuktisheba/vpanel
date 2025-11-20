import React from "react";
export interface Step {
  title: string;
  description?: string;
  hasError: boolean; // optional per-step error
}

interface ProjectProgressProps {
  steps: Step[];
  currentStep: number;
  // onStepChange?: (newStep: number) => void; // optional callback if parent wants to control step changes
}

const ProjectProgress: React.FC<ProjectProgressProps> = ({
  steps,
  currentStep,
  // onStepChange,
}) => {
  // const nextStep = () => {
  //   if (currentStep < steps.length) onStepChange?.(currentStep + 1);
  // };

  // const prevStep = () => {
  //   if (currentStep > 0) onStepChange?.(currentStep - 1);
  // };

  return (
    <div className="mt-4 mx-auto max-w-5xl space-y-8">
      {/* --- COMPONENT START: Progress Indicator Card --- */}
      <div className="rounded-sm border border-stroke bg-white shadow-default dark:border-strokedark dark:bg-boxdark">
        <div className="p-4 sm:p-6 xl:p-10">
          <div className="relative flex flex-col md:flex-row items-start md:items-center justify-between w-full">
            {/* Background Line (Absolute positioned to sit behind circles) */}
            {/* Only visible on desktop to avoid vertical line mess on mobile */}
            <div className="absolute top-4 left-0 hidden h-1 w-full -translate-y-1/2 bg-slate-200 md:block"></div>

            {/* Active Progress Line (Dynamic Width) */}
            <div
              className={`absolute top-4 left-0 hidden h-1 -translate-y-1/2 transition-all duration-500 ease-in-out md:block ${
                steps.slice(0, currentStep).some((step) => step.hasError)
                  ? "bg-red-500"
                  : "bg-indigo-500"
              }`}
              style={{
                width: `${
                  currentStep === 0 ? 0 : (currentStep / steps.length) * 100
                }%`,
              }}
            ></div>

            {steps.map((step: Step, index: number) => {
              // Determine state
              const isCompleted = index < currentStep;
              const isCurrent = index === currentStep;
              const isPending = index > currentStep;

              return (
                <div
                  key={index}
                  className={`relative z-10 flex flex-1 flex-row md:flex-col items-center md:justify-center gap-4 md:gap-0 ${
                    index !== 0 ? "mt-6 md:mt-0" : ""
                  }`}
                >
                  {/* Mobile Vertical Line (Connects steps visually on mobile) */}
                  {index !== steps.length - 1 && (
                    <div
                      className={`absolute left-4 top-8 h-full w-0.5 -translate-x-1/2 md:hidden ${
                        isCompleted ? "bg-indigo-500" : "bg-slate-200"
                      }`}
                    ></div>
                  )}

                  {/* The Circle Indicator */}
                  <div
                    className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2 transition-all duration-300
                      ${step.hasError ? "border-red-500 bg-red-100" : ""}
                      ${
                        !step.hasError && isCompleted
                          ? "border-indigo-500 bg-indigo-500"
                          : ""
                      }
                      ${
                        !step.hasError && isCurrent
                          ? "border-indigo-500 bg-white"
                          : ""
                      }
                      ${isPending ? "border-slate-200 bg-white" : ""}
                    `}
                  >
                    {
                      isCompleted &&
                        (step.hasError ? (
                          <div className="h-2.5 w-2.5 rounded-full bg-red-500" /> // e.g., red dot if error
                        ) : (
                          <div className="h-2.5 w-2.5 rounded-full bg-white" />
                        )) // normal completed dot
                    }

                    {isCurrent && (
                      <div
                        className={`h-2.5 w-2.5 rounded-full bg-indigo-500 ${
                          steps
                            .slice(0, currentStep)
                            .some((step) => step.hasError)
                            ? "bg-slate-200"
                            : "border-2 border-t-indigo-500 border-r-indigo-500 border-b-transparent border-l-transparent rounded-full animate-spin"
                        }`}
                      />
                    )}

                    {isPending && (
                      <div className="h-2.5 w-2.5 rounded-full bg-slate-200" />
                    )}
                  </div>

                  {/* Text Content */}
                  <div className="flex flex-col md:items-center md:pt-4 md:text-center">
                    <span
                      className={`text-sm font-bold transition-colors duration-300 ${
                        isCurrent || isCompleted
                          ? "text-black dark:text-white"
                          : "text-slate-500"
                      }`}
                    >
                      {step.title}
                    </span>
                    <span className="text-xs font-medium text-slate-500 dark:text-slate-400 mt-0.5 max-w-[120px]">
                      {step.description}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
      {/* --- COMPONENT END --- */}

      {/* Controls for Demo */}
      {/* <div className="flex justify-center gap-4">
        <button
          onClick={prevStep}
          disabled={currentStep === 0}
          className="flex justify-center rounded bg-slate-200 px-6 py-2 font-medium text-black hover:bg-opacity-90 disabled:opacity-50"
        >
          Previous
        </button>
        <button
          onClick={nextStep}
          disabled={currentStep === steps.length}
          className="flex justify-center rounded bg-indigo-500 px-6 py-2 font-medium text-gray hover:bg-opacity-90 disabled:opacity-50 text-white"
        >
          Next Step
        </button>
      </div> */}
    </div>
  );
};

export default ProjectProgress;
