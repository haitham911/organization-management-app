import { Metadata } from "next";
import Image from "next/image";
import { LinkedInLogoIcon } from "@radix-ui/react-icons";
import { UserAuthForm } from "@/components/auth/useAuthForm";

export const metadata: Metadata = {
  title: "User Management - Login",
  description: "User Management is a better way to handle authentication",
};

export default function LoginPage() {
  return (
    <div className="rbg-background overflow-hidden ">
      <div className="md:hidden">
        <Image
          src="/img-architecting-choices.jpg"
          width={1280}
          height={843}
          alt="User Management App"
          className="block dark:hidden"
        />
        <Image
          src="/img-architecting-choices.jpg"
          width={1280}
          height={843}
          alt="User Management App"
          className="hidden dark:block"
        />
      </div>
      <div className="container relative hidden h-screen flex-col items-center justify-center md:grid lg:max-w-none lg:grid-cols-2 lg:px-0">
        <div className="relative hidden h-full flex-col bg-muted p-10 dark:border-r lg:flex">
          <div className="auth-cover absolute inset-0" />
          <div className="relative z-20 flex items-center text-4xl font-medium">
            <Image
              src="/AL_logo_black.svg"
              width={250}
              height={250}
              alt="User Management App"
              className="dark:hidden"
            />
          </div>
          <div className="relative z-20 mt-auto">
            <blockquote className="space-y-2">
              <p className="text-lg text-white">
                User Management is a better way to handle authentication
              </p>
              <footer className="text-sm">
                <a href="#" target="_blank" rel="noopener noreferrer">
                  <LinkedInLogoIcon width={24} height={24} />
                </a>
              </footer>
            </blockquote>
          </div>
        </div>
        <div className="lg:p-8">
          <div className="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]">
            <div className="flex flex-col space-y-2 text-center">
              <h1 className="text-2xl font-semibold tracking-tight">Sign In</h1>
              <p className="text-sm text-muted-foreground">
                Login to your account
              </p>
            </div>
            <UserAuthForm />
          </div>
        </div>
      </div>
    </div>
  );
}
