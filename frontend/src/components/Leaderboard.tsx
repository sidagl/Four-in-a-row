import { useEffect, useState } from "react";
import axios from "axios";

const Leaderboard = ({ backendUrl }: { backendUrl: string }) => {
  const [data, setData] = useState<any[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await axios.get(`${backendUrl}/leaderboard`);
        setData(res.data);
      } catch (err) {
        console.error("Leaderboard fetch failed:", err);
      }
    };
    fetchData();
  }, [backendUrl]);

  return (
    <div className="space-y-2">
      {data.length > 0 ? (
        data.map((p, i) => (
          <div
            key={i}
            className="flex justify-between bg-white/10 px-3 py-2 rounded-lg"
          >
            <span>{p.username}</span>
            <span className="font-semibold">{p.wins}</span>
          </div>
        ))
      ) : (
        <p className="text-gray-400 text-sm text-center">No data yet</p>
      )}
    </div>
  );
};

export default Leaderboard;
