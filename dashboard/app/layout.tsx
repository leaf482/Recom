import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "EchoRec Dashboard",
  description: "Real-time music recommendation platform dashboard",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
