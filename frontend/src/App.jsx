import { useCallback, useEffect, useState } from "react";
import EnqueueForm from "./components/EnqueueForm";
import JobsTable from "./components/JobsTable";
import Summary from "./components/Summary";
import { apiUrl } from "./api";
import "./App.css";

const POLL_MS = 5000;

export default function App() {
  const [tab, setTab] = useState("jobs");
  const [jobs, setJobs] = useState([]);
  const [deadJobs, setDeadJobs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);

  const refetch = useCallback(async () => {
    try {
      const [jobsRes, deadRes] = await Promise.all([
        fetch(apiUrl("/jobs")),
        fetch(apiUrl("/dead-jobs")),
      ]);
      const jobsData = await jobsRes.json();
      const deadData = await deadRes.json();
      setJobs(Array.isArray(jobsData) ? jobsData : []);
      setDeadJobs(Array.isArray(deadData) ? deadData : []);
    } catch {
      // backend may be restarting; keep last known data
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refetch();
    const t = setInterval(refetch, POLL_MS);
    return () => clearInterval(t);
  }, [refetch]);

  const showToast = (msg, kind = "info") => {
    setToast({ msg, kind });
    setTimeout(() => setToast(null), 3000);
  };

  return (
    <div className="text-left">
      <header className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">GoJobberX</h1>
        <button
          onClick={refetch}
          className="text-sm border rounded px-3 py-1 hover:bg-gray-50"
        >
          refresh
        </button>
      </header>

      <EnqueueForm
        onEnqueued={(id) => {
          showToast(`enqueued ${id.slice(0, 8)}`, "ok");
          refetch();
        }}
        onError={(msg) => showToast(msg, "err")}
      />

      {toast && (
        <div
          className={`mb-4 px-3 py-2 rounded text-sm ${
            toast.kind === "err"
              ? "bg-red-50 border border-red-200 text-red-800"
              : toast.kind === "ok"
              ? "bg-green-50 border border-green-200 text-green-800"
              : "bg-gray-50 border border-gray-200 text-gray-800"
          }`}
        >
          {toast.msg}
        </div>
      )}

      <Summary jobs={jobs} />

      <div className="flex gap-2 mb-3 border-b">
        <button
          onClick={() => setTab("jobs")}
          className={`px-3 py-2 text-sm border-b-2 -mb-px ${
            tab === "jobs"
              ? "border-indigo-600 text-indigo-700 font-medium"
              : "border-transparent text-gray-600 hover:text-gray-900"
          }`}
        >
          jobs ({jobs.length})
        </button>
        <button
          onClick={() => setTab("dlq")}
          className={`px-3 py-2 text-sm border-b-2 -mb-px ${
            tab === "dlq"
              ? "border-indigo-600 text-indigo-700 font-medium"
              : "border-transparent text-gray-600 hover:text-gray-900"
          }`}
        >
          dead jobs ({deadJobs.length})
        </button>
      </div>

      {tab === "jobs" ? (
        <JobsTable jobs={jobs} loading={loading} />
      ) : (
        <JobsTable jobs={deadJobs} loading={loading} dlq />
      )}
    </div>
  );
}
