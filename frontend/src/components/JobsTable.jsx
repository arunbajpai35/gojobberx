import StatusBadge from "./StatusBadge";

export default function JobsTable({ jobs, loading, dlq }) {
  if (loading && jobs.length === 0) {
    return <div className="text-gray-500 text-sm py-8 text-center">loading...</div>;
  }
  if (jobs.length === 0) {
    return (
      <div className="text-gray-500 text-sm py-8 text-center border rounded bg-white">
        {dlq ? "no dead jobs" : "no jobs yet — enqueue one above"}
      </div>
    );
  }

  return (
    <div className="overflow-x-auto border rounded bg-white">
      <table className="w-full text-sm">
        <thead className="bg-gray-50 text-gray-600">
          <tr>
            <th className="px-3 py-2 text-left font-medium">id</th>
            <th className="px-3 py-2 text-left font-medium">payload</th>
            <th className="px-3 py-2 text-left font-medium">type</th>
            <th className="px-3 py-2 text-left font-medium">priority</th>
            {!dlq && <th className="px-3 py-2 text-left font-medium">status</th>}
            <th className="px-3 py-2 text-left font-medium">retries</th>
            <th className="px-3 py-2 text-left font-medium">
              {dlq ? "failed at" : "updated"}
            </th>
          </tr>
        </thead>
        <tbody>
          {jobs.map((j) => (
            <tr key={j.id} className="border-t hover:bg-gray-50">
              <td className="px-3 py-2 font-mono text-xs text-gray-600">
                {String(j.id).slice(0, 8)}
              </td>
              <td className="px-3 py-2">{j.payload}</td>
              <td className="px-3 py-2">{j.type}</td>
              <td className="px-3 py-2 capitalize">{j.priority}</td>
              {!dlq && (
                <td className="px-3 py-2">
                  <StatusBadge status={j.status} />
                </td>
              )}
              <td className="px-3 py-2">{j.retries}</td>
              <td className="px-3 py-2 text-gray-500 text-xs">
                {new Date(dlq ? j.failed_at : j.updated_at).toLocaleTimeString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
