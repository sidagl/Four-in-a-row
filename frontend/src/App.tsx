import { useState } from "react";
import useWebSocket from "react-use-websocket";
import { motion } from "framer-motion";
import GameBoard from "./components/GameBoard";
import Leaderboard from "./components/Leaderboard";
import "./index.css";

const backendUrl = import.meta.env.VITE_BACKEND_URL || "http://localhost:9090";
const WS_URL = backendUrl.replace("http", "ws") + "/ws";

console.log("🧠 Backend URL =>", backendUrl);
console.log("🧩 WS_URL =>", WS_URL);

function App() {
  const [username, setUsername] = useState("");
  const [connected, setConnected] = useState(false);
  const [waiting, setWaiting] = useState(true);
  const [messages, setMessages] = useState<any[]>([]);
  const [socketUrl, setSocketUrl] = useState("");
  const [board, setBoard] = useState<number[][]>([]);
  const [turn, setTurn] = useState<number>(0);

  const shouldConnect = Boolean(socketUrl);

  // ✅ useWebSocket with stable heartbeat and reconnection
  const { sendJsonMessage } = useWebSocket(
    socketUrl,
    {
      share: false,
      shouldReconnect: () => true,
      reconnectAttempts: 10,
      reconnectInterval: 3000,

      heartbeat: {
        message: JSON.stringify({ type: "ping" }),
        returnMessage: "pong",
        interval: 45000, // every 45s
      },

      onOpen: () => {
        console.log("✅ Connected to server");
        setConnected(true);
      },

      onClose: (e) => {
        console.warn("⚠️ WebSocket closed:", e.reason || e.code);
        setConnected(false);
        setWaiting(true);
      },

      onError: (err) => {
        console.error("❌ WebSocket error:", err);
      },

      onMessage: (msg) => {
        console.log("📩 Raw message:", msg.data);

        // 🧩 Split concatenated JSONs safely (handles back-to-back JSON)
        const jsonParts: string[] = msg.data
          .split("}{")
          .map((part: string, idx: number, arr: string[]) =>
            idx < arr.length - 1 ? part + "}" : part
          )
          .map((p: string, i: number) => (i > 0 ? "{" + p : p));

        jsonParts.forEach((raw: string) => {
          let data: any;
          try {
            data = JSON.parse(raw);
          } catch {
            console.warn("⚠️ Invalid JSON part:", raw);
            return;
          }

          console.log("✅ Parsed JSON:", data);
          setMessages((prev) => [...prev, data]);

          switch (data.type) {
            case "ping":
              // Respond to backend heartbeat
              sendJsonMessage({ type: "pong" });
              console.log("💓 Sent pong response");
              break;

            case "start":
            case "start_game":
              console.log("🎮 Game start detected!");
              setWaiting(false);
              break;

            case "state":
              setBoard(data.board || []);
              setTurn(data.turn || 0);
              console.log("🧩 Board updated:", data.board);
              break;

            case "end":
              alert(`🏁 Game Over! Winner: ${data.winner}`);
              setWaiting(true);
              break;

            default:
              console.log("ℹ️ Unknown message type:", data.type);
          }
        });
      },
    },
    shouldConnect
  );

  // 🔗 Connect when username entered
  const connect = () => {
    if (!username.trim()) return alert("Enter username first!");
    console.log(`🔗 Connecting as ${username}`);
    setSocketUrl(`${WS_URL}?username=${username}`);
  };

  // 🎯 Handle player moves
  const handleMove = (col: number) => {
    if (!connected || waiting) return;
    sendJsonMessage({ type: "move", column: col });
  };

  // 🎨 UI Rendering
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.8 }}
      className="flex flex-col md:flex-row gap-8 p-6 min-h-screen text-white bg-gradient-to-br from-gray-900 via-blue-950 to-gray-800"
    >
      {/* 🎮 Game Section */}
      <motion.div
        layout
        className="glass p-6 md:w-2/3 text-center rounded-2xl border border-gray-700 bg-opacity-30 backdrop-blur-md shadow-lg"
        transition={{ duration: 0.3 }}
      >
        {!connected ? (
          // Login screen
          <div className="flex flex-col items-center gap-4">
            <h1 className="text-4xl font-bold mb-4">🎮 4-in-a-Row</h1>
            <input
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="px-4 py-2 rounded-lg text-black w-64 outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={connect}
              className="px-5 py-2 bg-blue-600 rounded-lg text-white font-semibold hover:bg-blue-700 transition-all"
            >
              Start Game
            </button>
          </div>
        ) : waiting ? (
          // Waiting screen
          <div className="flex flex-col items-center justify-center h-80 text-lg animate-pulse">
            ⏳ Waiting for another player to join...
          </div>
        ) : (
          // Game Board
          <GameBoard
            messages={messages}
            handleMove={handleMove}
            board={board}
            turn={turn}
            username={username}
          />
        )}
      </motion.div>

      {/* 🏆 Leaderboard Section */}
      <motion.div
        layout
        className="glass p-6 md:w-1/3 rounded-2xl border border-gray-700 bg-opacity-30 backdrop-blur-md shadow-lg"
        initial={{ x: 50, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ duration: 0.5 }}
      >
        <h2 className="text-2xl font-semibold mb-4 text-center">
          🏆 Leaderboard
        </h2>
        <Leaderboard backendUrl={backendUrl} />
      </motion.div>
    </motion.div>
  );
}

export default App;
