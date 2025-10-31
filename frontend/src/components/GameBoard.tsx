import { useEffect, useState } from "react";

interface GameBoardProps {
  board: number[][];
  turn: number;
  handleMove: (col: number) => void;
  username: string;
}

export default function GameBoard({ board, turn, handleMove, username }: GameBoardProps) {
  const [localBoard, setLocalBoard] = useState<number[][]>(
    Array.from({ length: 6 }, () => Array(7).fill(0))
  );

  useEffect(() => {
    if (board && board.length) setLocalBoard(board);
  }, [board]);

  return (
    <div className="flex flex-col items-center justify-center mt-8 space-y-6">
      {/* 🎯 Turn Indicator */}
      <h2 className="text-3xl font-bold mb-2">
        🎯 Turn: Player {turn} {turn === 1 ? "🟡" : "🔴"}
      </h2>

      {/* 🎮 Board Container */}
      <div
        className="grid grid-cols-7 gap-3 bg-blue-700 p-6 rounded-2xl shadow-2xl border-4 border-blue-900"
        style={{
          width: "max-content",
        }}
      >
        {localBoard.map((row, rowIndex) =>
          row.map((cell, colIndex) => (
            <div
              key={`${rowIndex}-${colIndex}`}
              onClick={() => handleMove(colIndex)}
              className={`w-14 h-14 rounded-full border-4 border-blue-900 transition-all cursor-pointer transform hover:scale-110 ${
                cell === 1
                  ? "bg-yellow-400"
                  : cell === 2
                  ? "bg-red-500"
                  : "bg-gray-200"
              }`}
            ></div>
          ))
        )}
      </div>

      {/* 👤 Player Info */}
      <p className="text-gray-300 text-sm">
        You are playing as <span className="font-semibold">{username}</span>
      </p>
    </div>
  );
}
