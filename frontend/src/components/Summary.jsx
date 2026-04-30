const order = ["queued", "processing", "completed", "failed"];
const colors = {
  queued: "bg-yellow-50 border-yellow-200 text-yellow-800",
  processing: "bg-blue-50 border-blue-200 text-blue-800",
  completed: "bg-green-50 border-green-200 text-green-800",
  failed: "bg-red-50 border-red-200 text-red-800",
};

export default function Summary({ jobs }) {
  const counts = order.reduce((acc, s) => ({ ...acc, [s]: 0 }), {});
  for (const j of jobs) {
    if (counts[j.status] !== undefined) counts[j.status]++;
  }

  return (
    <div className="flex gap-3 mb-4">
      {order.map((s) => (
        <div
          key={s}
          className={`px-3 py-2 rounded border text-sm ${colors[s]}`}
        >
          <div className="text-xs uppercase tracking-wide">{s}</div>
          <div className="text-lg font-semibold">{counts[s]}</div>
        </div>
      ))}
    </div>
  );
}
