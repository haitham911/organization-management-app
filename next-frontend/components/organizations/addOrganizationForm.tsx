"use client";

import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  NewOrganizationForm,
  newOrganizationSchema,
} from "@/lib/schemas/organizations/newOrganizationSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { Label } from "../ui/label";
import { Input } from "../ui/input";
import { Button } from "../ui/button";
import { AppRoutes } from "@/config/routes";
import { Alert } from "../ui/alert";
import { useState } from "react";
import { TNewOrganization } from "@/lib/types/organizationTypes";
import { newOrganization } from "@/lib/services/organizations/organizationsApi";
import { toast } from "sonner";

export const AddOrganizationForm = () => {
  const router = useRouter();
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<NewOrganizationForm>({
    resolver: zodResolver(newOrganizationSchema),
    defaultValues: {
      name: "",
      email: "",
    },
  });

  const submitFormHandler = async (data: TNewOrganization) => {
    try {
      setServerError(null); // Clear any previous errors
      const response = await newOrganization(data);
      if (response.ok) {
        toast.success("Organization created successfully");
      }
    } catch (error: any) {
      setServerError("An unexpected error occurred");
    }
  };

  const backToListHandler = () => {
    router.refresh();
    router.push(AppRoutes.organizations.list);
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Create Organization</CardTitle>
      </CardHeader>
      <CardContent>
        {serverError && (
          <div className="mb-2 flex justify-center rounded-md bg-red-300 p-2">
            <p className="text-md capitalize text-red-900">{serverError}</p>
          </div>
        )}
        <form onSubmit={handleSubmit(submitFormHandler)}>
          <div className="grid w-full items-center gap-4">
            <div className="flex flex-col space-y-1.5">
              {errors.name && (
                <Alert variant="destructive">
                  <p>{errors.name.message}</p>
                </Alert>
              )}
              <Label htmlFor="name">Name</Label>
              <Input
                {...register("name")}
                id="name"
                placeholder="Name of your organization"
                disabled={isSubmitting}
              />
            </div>
            <div className="flex flex-col space-y-1.5">
              {errors.email && (
                <Alert variant="destructive">
                  <p>{errors.email.message}</p>
                </Alert>
              )}
              <Label htmlFor="email">Email</Label>
              <Input
                {...register("email")}
                id="email"
                placeholder="name@example.com"
                type="email"
                autoCapitalize="none"
                autoComplete="email"
                autoCorrect="off"
                disabled={isSubmitting}
              />
            </div>
            <div>
              <Button type="submit" disabled={isSubmitting}>
                Create Organization
              </Button>
            </div>
          </div>
        </form>
      </CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="outline" onClick={backToListHandler}>
          Back to list
        </Button>
      </CardFooter>
    </Card>
  );
};
