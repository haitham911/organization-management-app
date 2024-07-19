import { ProductsList } from "@/components/products/productList";
import { getProducts } from "@/lib/services/products/productsApi";

export default async function Products() {
  const products = await getProducts();
  console.log("products", products);
  return (
    <section>
      <h1>Products</h1>
      <ProductsList />
    </section>
  );
}
