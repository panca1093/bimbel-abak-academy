import type { Metadata } from "next";
import { Public_Sans, Source_Serif_4, Spline_Sans_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";

const publicSans = Public_Sans({ subsets: ["latin"], variable: "--font-public-sans" });
const sourceSerif = Source_Serif_4({ subsets: ["latin"], variable: "--font-source-serif" });
const splineMono = Spline_Sans_Mono({ subsets: ["latin"], variable: "--font-spline-mono" });

export const metadata: Metadata = {
  title: "Abak Academy",
  description: "Tingkatkan Skills Bersama Abak Academy",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      <body
        className={`${publicSans.variable} ${sourceSerif.variable} ${splineMono.variable} antialiased`}
      >
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
