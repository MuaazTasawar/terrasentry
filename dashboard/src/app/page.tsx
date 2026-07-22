import { redirect } from "next/navigation";

// The dashboard has no standalone home view — scans is the default landing
// page, matching what an on-call engineer wants to see first.
export default function RootPage() {
  redirect("/scans");
}