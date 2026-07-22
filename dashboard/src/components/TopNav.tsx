"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { clearToken, getToken } from "@/lib/apiClient";

const LINKS = [
  { href: "/scans", label: "Scans" },
  { href: "/drift", label: "Drift" },
  { href: "/audit", label: "Audit" },
];

export default function TopNav() {
  const pathname = usePathname();
  const router = useRouter();
  const [loggedIn, setLoggedIn] = useState(false);

  useEffect(() => {
    // Syncing local state with localStorage (an external system) on route
    // change. See https://react.dev/learn/you-might-not-need-an-effect
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoggedIn(Boolean(getToken()));
  }, [pathname]);

  function handleLogout() {
    clearToken();
    router.push("/login");
  }

  return (
    <header className="border-b border-zinc-800 bg-zinc-950/80 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href="/scans" className="flex items-center gap-2 font-mono text-sm font-semibold text-zinc-100">
          <span className="inline-block h-2 w-2 rounded-full bg-emerald-500" />
          TerraSentry
        </Link>

        <nav className="flex items-center gap-1">
          {LINKS.map((link) => {
            const active = pathname?.startsWith(link.href);
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`rounded-sm px-3 py-1.5 text-sm transition-colors ${
                  active
                    ? "bg-zinc-800 text-zinc-100"
                    : "text-zinc-400 hover:text-zinc-100"
                }`}
              >
                {link.label}
              </Link>
            );
          })}

          {loggedIn ? (
            <button
              onClick={handleLogout}
              className="ml-3 rounded-sm px-3 py-1.5 text-sm text-zinc-500 hover:text-rose-400"
            >
              Log out
            </button>
          ) : (
            <Link
              href="/login"
              className="ml-3 rounded-sm bg-zinc-100 px-3 py-1.5 text-sm font-medium text-zinc-900 hover:bg-white"
            >
              Log in
            </Link>
          )}
        </nav>
      </div>
    </header>
  );
}