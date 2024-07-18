import { Icons } from "@/components/ui/icons";
import dynamic from "next/dynamic";

const NewOrganizationForm = dynamic(
  () =>
    import("@/components/organizations/addOrganizationForm").then(
      (mod) => mod.AddOrganizationForm
    ),
  {
    ssr: false,
    loading: () => <Icons.spinner />,
  }
);

export default function Organizations() {
  return (
    <section className="p-10 flex flex-col gap-4">
      <h1 className="font-bold text-xxl">New Organization</h1>
      <NewOrganizationForm />
    </section>
  );
}
