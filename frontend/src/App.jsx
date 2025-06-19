import React, { useEffect, useState } from "react";

function App() {
    const [jobs, setJobs] = useState([]);

    useEffect(() => {
        const fetchJobs = async () => {
            try {
                const res = await fetch("http://localhost:8080/jobs");
                const data = await res.json();
                setJobs(data);
            } catch (err) {
                console.error("Failed to fetch jobs:", err);
            }
        };

        fetchJobs();
        const interval = setInterval(fetchJobs, 5000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="p-6">
            <h1 className="text-2xl font-bold mb-4">📊 GoJobberX Dashboard</h1>
            <table className="w-full border-collapse text-sm">
                <thead className="bg-gray-100">
                <tr>
                    <th className="px-4 py-2 text-left">Job ID</th>
                    <th className="px-4 py-2 text-left">Payload</th>
                    <th className="px-4 py-2 text-left">Type</th>
                    <th className="px-4 py-2 text-left">Priority</th>
                    <th className="px-4 py-2 text-left">Status</th>
                    <th className="px-4 py-2 text-left">Retries</th>
                </tr>
                </thead>
                <tbody>
                {jobs.length === 0 ? (
                    <tr>
                        <td colSpan="6" className="px-4 py-2 text-center text-gray-500">
                            No jobs found.
                        </td>
                    </tr>
                ) : (
                    jobs.map((job) => (
                        <tr key={job.id} className="border-t hover:bg-gray-50">
                            <td className="px-4 py-2 font-mono text-xs">{job.id}</td>
                            <td className="px-4 py-2">{job.payload}</td>
                            <td className="px-4 py-2">{job.type}</td>
                            <td className="px-4 py-2 capitalize">{job.priority}</td>
                            <td className="px-4 py-2">{job.status}</td>
                            <td className="px-4 py-2">{job.retries}</td>
                        </tr>
                    ))
                )}
                </tbody>
            </table>
        </div>
    );
}

export default App;
