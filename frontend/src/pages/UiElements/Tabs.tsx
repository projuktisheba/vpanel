import { useState, FC, ReactNode } from "react";

interface Tab {
  label: string;
  content: ReactNode;
}

interface PillTabsProps {
  tabs: Tab[];
  defaultIndex?: number;
}

const PillTabs: FC<PillTabsProps> = ({ tabs, defaultIndex = 0 }) => {
  const [activeIndex, setActiveIndex] = useState(defaultIndex);

  return (
    <div>
      {/* Tab buttons */}
      <div className="inline-flex bg-gray-200 rounded p-0.5 dark:bg-gray-700">
        {tabs.map((tab, index) => (
          <button
            key={index}
            onClick={() => setActiveIndex(index)}
            className={`px-4 py-2 text-sm font-medium rounded transition-colors
              ${activeIndex === index 
                ? "bg-white text-black shadow-sm dark:bg-gray-800 dark:text-white"
                : "text-gray-500 hover:text-black dark:text-gray-400 dark:hover:text-white"
              }
            `}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Active tab content */}
      <div className="mt-4">
        {tabs[activeIndex]?.content}
      </div>
    </div>
  );
};

export default PillTabs;
