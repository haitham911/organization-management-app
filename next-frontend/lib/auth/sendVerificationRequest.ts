import { Theme } from 'next-auth';
import { EmailConfig } from 'next-auth/providers/email';
import nodemailer from 'nodemailer';

// Define types for the function parameters

type TVerificationRequestParams = {
  identifier: string;
  url: string;
  expires: Date;
  token: string;
  theme?: Theme;
  provider: EmailConfig;
}

// Function to send verification request
export async function sendVerificationRequest({ identifier: email, url, provider }: TVerificationRequestParams) {
  const { server, from } = provider;
  const transport = nodemailer.createTransport(server);
  console.log(`Sending email to ${email}`);
  const result:any = await transport.sendMail({
    to: email,
    from,
    subject: "Sign in to your account",
    text: `Sign in to your account by clicking the link: ${url}`,
    html: `<p>Sign in to your account by clicking the link below:</p><p><a href="${url}">CLICK HERE</a></p>`
  });

  if (result.rejected.length) {
    throw new Error('Email could not be sent');
  }
}
