import { buttonVariants } from "@/components/ui/button";
import { AppRoutes } from "@/config/routes";
import { getOrganizationsList } from "@/lib/services/organizations/organizationsApi";
import { cn } from "@/lib/utils";
import Link from "next/link";

import dynamic from "next/dynamic";
import { Icons } from "@/components/ui/icons";
const OrganizationList = dynamic(
  () =>
    import("@/components/organizations/organizationList").then(
      (mod) => mod.OrganizationList
    ),
  {
    ssr: false,
    loading: () => <Icons.spinner />,
  }
);

export default async function Organizations() {
  const organizationsLists = await getOrganizationsList();
  return (
    <section className="p-5">
      <h1 className="text-xl font-bold">Organizations</h1>
      <div className="flex flex-col gap-4">
        <div>
          <Link
            className={cn(buttonVariants({ variant: "destructive" }))}
            href={AppRoutes.organizations.new}
          >
            New Organization
          </Link>
        </div>
        <OrganizationList organizations={organizationsLists} />
      </div>
    </section>
  );
}
