type MainTab = "go-crud";

interface TabNavigationProps {
  activeTab: MainTab;
  onTabChange: (tab: MainTab) => void;
}

// Add a new entry here (and a matching value to MainTab above) when the next
// chapter's tab is ready — e.g. { id: "rust-crud", label: "🦀 Rust" }.
const TABS: { id: MainTab; label: string }[] = [
  { id: "go-crud", label: "🐹 Go" },
];

export default function TabNavigation({
  activeTab,
  onTabChange,
}: TabNavigationProps) {
  return (
    <nav className="bg-white dark:bg-zinc-800 border-b border-gray-200 dark:border-zinc-700">
      <div className="max-w-7xl mx-auto px-6">
        <div className="flex gap-0">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => onTabChange(tab.id)}
              className={`px-6 py-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? "border-blue-500 text-blue-600 dark:text-blue-400"
                  : "border-transparent text-gray-500 dark:text-zinc-400 hover:text-gray-700 dark:hover:text-zinc-200"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>
    </nav>
  );
}
