import jwt, { SignOptions } from 'jsonwebtoken';

export const generateToken = (email: string): string => {
  const secretKey = process.env.SECRET as string; // Ensure this is set in your environment
  if (!secretKey) {
    throw new Error('Secret key is not defined in environment variables');
  }
  const payload = { email };
  const options: SignOptions = {
    expiresIn: '1h', // Token expiration time
  };

  const token = jwt.sign(payload, secretKey, options);
  return token;
};
