"use client";

import { useEffect, useState } from "react";
import { API_BASE } from "@/lib/api";

type Health = { status: string; postgres: string; redis: string };

export default function StudentDashboard() {
  const [health, setHealth] = useState<Health | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(`${API_BASE}/health`)
      .then((r) => r.json())
      .then(setHealth)
      .catch((e) => setError(String(e)));
  }, []);

  return (
    <main className="min-h-screen bg-brand-50 px-6 py-12">
      <div className="mx-auto max-w-2xl">
        <h1 className="font-serif text-3xl font-bold text-brand-900">Abak Academy</h1>
        <p className="mt-2 text-brand-700">Student shell — scaffold placeholder.</p>

        <div className="mt-8 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-brand-600">
            API health
          </h2>
          {error && <p className="mt-2 font-mono text-sm text-red-600">{error}</p>}
          {health ? (
            <ul className="mt-2 space-y-1 font-mono text-sm text-gray-700">
              <li>status: {health.status}</li>
              <li>postgres: {health.postgres}</li>
              <li>redis: {health.redis}</li>
            </ul>
          ) : (
            !error && <p className="mt-2 text-sm text-gray-500">checking…</p>
          )}
        </div>
      </div>
    </main>
  );
}
