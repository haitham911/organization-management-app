/* eslint-disable */
export default function MemoryAdapter() {
    let users:any[] = [];
    let sessions:any = [];
    let verificationTokens:any = [];
  
    return {
      async createUser(user:any) {
        users.push(user);
        return user;
      },
      async getUser(id:any) {
        return users.find(user => user.id === id) || null;
      },
      async getUserByEmail(email:string) {
        return users.find(user => user.email === email) || null;
      },
      async getUserByAccount({ providerAccountId, provider }:any) {
        return users.find(user => user.provider === provider && user.providerAccountId === providerAccountId) || null;
      },
      async updateUser(user:any) {
        users = users.map(u => (u.id === user.id ? user : u));
        return user;
      },
      async deleteUser(userId:any) {
        users = users.filter(user => user.id !== userId);
      },
      async linkAccount(account:any) {
        const user = users.find(user => user.id === account.userId);
        if (user) {
          user.accounts = user.accounts || [];
          user.accounts.push(account);
        }
      },
      async unlinkAccount({ providerAccountId, provider }:any) {
        const user = users.find(user => user.provider === provider && user.providerAccountId === providerAccountId);
        if (user) {
          user.accounts = user.accounts.filter((account:any) => account.provider !== provider || account.providerAccountId !== providerAccountId);
        }
      },
      async createSession(session:any) {
        sessions.push(session);
        return session;
      },
      async getSessionAndUser(sessionToken:any) {
        const session = sessions.find((session:any) => session.sessionToken === sessionToken);
        if (!session) return null;
        const user = users.find(user => user.id === session.userId);
        return { session, user };
      },
      async updateSession(session:any) {
        sessions = sessions.map((s:any) => (s.sessionToken === session.sessionToken ? session : s));
        return session;
      },
      async deleteSession(sessionToken:any) {
        sessions = sessions.filter((session:any) => session.sessionToken !== sessionToken);
      },
      async createVerificationToken(token:any) {
        verificationTokens.push(token);
        return token;
      },
      async useVerificationToken({ identifier, token }:any) {
        const foundToken = verificationTokens.find((v:any) => v.identifier === identifier && v.token === token);
        if (!foundToken) return null;
        verificationTokens = verificationTokens.filter((v:any) => v.identifier !== identifier || v.token !== token);
        return foundToken;
      },
    };
  }
  