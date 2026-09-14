"use client";

import { useState } from "react";
import PageHeader from "@/app/components/PageHeader";
import TabNavigation from "@/app/components/TabNavigation";
import GoCrud from "@/app/components/GoCrud";

type MainTab = "go-crud";

export default function Home() {
  const [mainTab, setMainTab] = useState<MainTab>("go-crud");

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-zinc-900 font-sans">
      <PageHeader />
      <TabNavigation activeTab={mainTab} onTabChange={(tab) => setMainTab(tab)} />
      <div className={mainTab === "go-crud" ? undefined : "hidden"}>
        <GoCrud />
      </div>
    </div>
  );
}
