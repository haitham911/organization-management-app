import type { Metadata } from "next";
import { Inter as FontSans } from "next/font/google";
import "./globals.css";

const fontSans = FontSans({
  subsets: ["latin"],
  variable: "--font-sans",
});
import { cn } from "@/lib/utils";
import { Toaster } from "@/components/ui/sonner";
import AuthProvider from "@/lib/auth/authProvider";

export const metadata: Metadata = {
  title: "User Management",
  description: "Generated by create next app",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <AuthProvider>
        {" "}
        <body
          className={cn(
            "min-h-screen bg-background font-sans antialiased",
            fontSans.variable
          )}
        >
          <Toaster />
          <main className="bg-gray-50 h-screen">{children}</main>
        </body>
      </AuthProvider>
    </html>
  );
}
