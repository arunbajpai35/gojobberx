const base = import.meta.env.VITE_API_BASE || "";

export const apiUrl = (path) => `${base}${path}`;
