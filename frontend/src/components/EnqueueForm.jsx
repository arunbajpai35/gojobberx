import { useState } from "react";

const initial = {
  payload: "",
  type: "send_email",
  duration: 1,
  priority: "medium",
};

export default function EnqueueForm({ onEnqueued, onError }) {
  const [form, setForm] = useState(initial);
  const [submitting, setSubmitting] = useState(false);

  const update = (k) => (e) => {
    const v = k === "duration" ? Number(e.target.value) : e.target.value;
    setForm({ ...form, [k]: v });
  };

  const submit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const res = await fetch("/job", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      const data = await res.json();
      if (!res.ok) {
        onError?.(data.error || `error ${res.status}`);
      } else {
        onEnqueued?.(data.job_id);
        setForm(initial);
      }
    } catch (err) {
      onError?.(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={submit}
      className="bg-white border rounded p-4 mb-6 grid grid-cols-1 md:grid-cols-5 gap-3 items-end"
    >
      <label className="flex flex-col text-sm md:col-span-2">
        <span className="text-gray-600 mb-1">payload</span>
        <input
          className="border rounded px-2 py-1"
          value={form.payload}
          onChange={update("payload")}
          placeholder="hello@example.com"
          required
        />
      </label>

      <label className="flex flex-col text-sm">
        <span className="text-gray-600 mb-1">type</span>
        <select
          className="border rounded px-2 py-1"
          value={form.type}
          onChange={update("type")}
        >
          <option value="send_email">send_email</option>
          <option value="generate_invoice">generate_invoice</option>
        </select>
      </label>

      <label className="flex flex-col text-sm">
        <span className="text-gray-600 mb-1">duration (s)</span>
        <input
          type="number"
          min={0}
          max={600}
          className="border rounded px-2 py-1"
          value={form.duration}
          onChange={update("duration")}
        />
      </label>

      <label className="flex flex-col text-sm">
        <span className="text-gray-600 mb-1">priority</span>
        <select
          className="border rounded px-2 py-1"
          value={form.priority}
          onChange={update("priority")}
        >
          <option value="high">high</option>
          <option value="medium">medium</option>
          <option value="low">low</option>
        </select>
      </label>

      <button
        type="submit"
        disabled={submitting}
        className="md:col-span-5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white rounded px-4 py-2 text-sm font-medium"
      >
        {submitting ? "enqueueing..." : "enqueue job"}
      </button>
    </form>
  );
}
