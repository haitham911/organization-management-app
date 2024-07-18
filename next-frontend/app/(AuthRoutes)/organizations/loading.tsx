import { Icons } from "@/components/ui/icons";

export default function Loading() {
  return (
    <section className="flex h-screen w-full items-center justify-center">
      <Icons.spinner className="h-10 w-10" />
      <p className="ml-2">Loading...</p>
    </section>
  );
}
