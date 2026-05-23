import { Routes, Route } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import McpIntegration from "./pages/McpIntegration";
import Servers from "./pages/Servers";
import Logs from "./pages/Logs";
import Monitor from "./pages/Monitor";
import OpenCodeSessions from "./pages/OpenCodeSessions";

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/mcp" element={<McpIntegration />} />
        <Route path="/servidores" element={<Servers />} />
        <Route path="/opencode" element={<OpenCodeSessions />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/monitor" element={<Monitor />} />
      </Route>
    </Routes>
  );
}
