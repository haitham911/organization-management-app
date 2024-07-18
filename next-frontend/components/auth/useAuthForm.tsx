"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Label } from "../ui/label";
import { Input } from "../ui/input";
import { Button } from "../ui/button";
import Link from "next/link";
import { Icons } from "../ui/icons";
import { useForm } from "react-hook-form";
import { LoginForm, LoginSchema } from "@/lib/schemas/auth/loginSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import { Alert } from "../ui/alert";

interface UserAuthFormProps extends React.HTMLAttributes<HTMLDivElement> {}

export function UserAuthForm({ className, ...props }: UserAuthFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(LoginSchema),
  });

  async function onSubmit(event: React.SyntheticEvent) {
    event.preventDefault();
  }

  return (
    <div className={cn("grid gap-6", className)} {...props}>
      {errors.email && (
        <Alert variant="destructive" className="text-sm">
          <p>{errors.email.message}</p>
        </Alert>
      )}
      <form onSubmit={onSubmit}>
        <div className="grid gap-2">
          <div className="grid gap-1">
            <Label className="sr-only" htmlFor="email">
              Email
            </Label>
            <Input
              id="email"
              placeholder="name@example.com"
              type="email"
              autoCapitalize="none"
              autoComplete="email"
              autoCorrect="off"
              disabled={isSubmitting}
            />
          </div>
          <Button disabled={isSubmitting}>
            {isSubmitting && (
              <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
            )}
            Sign In with Email
          </Button>
        </div>
      </form>
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-t" />
        </div>
        <div className="relative flex justify-center text-xs uppercase hidden">
          <span className="bg-background px-2 text-muted-foreground">
            Continue with
          </span>
        </div>
      </div>
      <Link href="/api/auth/login" className="w-full hidden">
        <Button
          variant="outline"
          type="button"
          disabled={isSubmitting}
          className="w-full"
        >
          {isSubmitting ? (
            <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Icons.auth0 className="mr-2 h-4 w-4" />
          )}{" "}
          Auth0
        </Button>
      </Link>
    </div>
  );
}
