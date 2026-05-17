import { Routes, Route } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import McpIntegration from "./pages/McpIntegration";
import Modules from "./pages/Modules";
import Servers from "./pages/Servers";
import Logs from "./pages/Logs";

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/mcp" element={<McpIntegration />} />
        <Route path="/modulos" element={<Modules />} />
        <Route path="/servidores" element={<Servers />} />
        <Route path="/logs" element={<Logs />} />
      </Route>
    </Routes>
  );
}
