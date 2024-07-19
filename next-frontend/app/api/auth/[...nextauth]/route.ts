import NextAuth from "next-auth";
import EmailProvider from "next-auth/providers/email";
import MemoryAdapter from "@/lib/auth/memoryAdapter";
import { sendVerificationRequest } from "@/lib/auth/sendVerificationRequest";
import { generateToken } from "@/lib/auth/jwtToken";
import { REDIRECT_AFTER_LOGIN } from "@/lib/auth/authConfig";

const handler = NextAuth({
  providers: [
    EmailProvider({
      server: {
        host: process.env.EMAIL_SERVER_HOST,
        port: Number(process.env.EMAIL_SERVER_PORT),
        auth: {
          user: process.env.EMAIL_SERVER_USER,
          pass: process.env.EMAIL_SERVER_PASSWORD,
        },
      },
      from: process.env.EMAIL_FROM,
      sendVerificationRequest,
    }),
  ],
  session: {
    strategy: "jwt",
  },
  jwt: {
    secret: process.env.SECRET,
  },
  adapter: MemoryAdapter(),
  callbacks: {
    async redirect({ url, baseUrl }) {
      // Ensure that the URL is relative to your site
      if (url.startsWith(baseUrl)) {
        return `${baseUrl}${REDIRECT_AFTER_LOGIN}`;
      } else if (url.startsWith("/")) {
        return `${baseUrl}${url}`;
      }
      return baseUrl;
    },
    async session({ session, token }) {
      if (session.user) {
        session.user.email = token.email;
      }
      return session;
    },
    async jwt({ token, user }: any) {
      console.log("JWT TOKEN", token)
        // Create a new session for the user
        const sessionToken = generateToken(token.email);
        console.log("GENERATED JWT", sessionToken);
        token.sessionToken = sessionToken;
        return token;
  
    },
  },
});
export { handler as GET, handler as POST };
