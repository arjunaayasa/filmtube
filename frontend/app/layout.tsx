import type { Metadata } from "next"
import { Be_Vietnam_Pro } from "next/font/google"
import Script from "next/script"
import "./globals.css"

const beVietnamPro = Be_Vietnam_Pro({
  subsets: ["latin"],
  variable: "--font-display",
  display: "swap",
})

export const metadata: Metadata = {
  title: "FilmTube - Indie Film Streaming Platform",
  description: "Discover and watch short films and feature films from emerging creators worldwide.",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <head>
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <link href="https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:wght@300;400;500;600;700;900&display=swap" rel="stylesheet" />
      </head>
      <body className={beVietnamPro.variable}>
        <Script src="https://cdn.jsdelivr.net/npm/hls.js@latest" strategy="beforeInteractive" />
        {children}
      </body>
    </html>
  )
}
