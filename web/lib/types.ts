export type ProductType = "book" | "course" | "package";

export type ProductStatus = "draft" | "published" | "hidden" | "archived";

export type OrderStatus =
  | "cart"
  | "payment_pending"
  | "paid"
  | "processing"
  | "shipped"
  | "completed"
  | "payment_expired"
  | "cancelled";

export type AdminOrderFilterStatus = "all" | "pending" | "paid" | "processing" | "shipped" | "failed" | "refunded";

export interface School {
  id: string;
  name: string;
  code?: string;
}

export interface User {
  id: string;
  email?: string;
  username?: string;
  name?: string;
  role?: string;
  school_id?: string;
  status?: string;
  otp_enabled?: boolean;
  phone?: string;
  nis?: string;
  grade?: string;
  target_exam?: string;
  alamat_domisili?: string;
  dob?: string;
  gender?: string;
  photo_url?: string;
  created_at?: string;
  updated_at?: string;
}

export interface LoginResponse {
  access_token?: string;
  refresh_token?: string;
  user?: User;
  otp_required?: boolean;
  pending_token?: string;
}

export interface Product {
  id: string;
  type: ProductType;
  name: string;
  description?: string;
  price: number;
  stock?: number;
  status?: ProductStatus;
  weight_grams?: number;
  image_url?: string;
  course_ids?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface AdminCreateProductInput {
  type: ProductType;
  name: string;
  description?: string;
  price: number;
  stock?: number;
  course_ids?: string[];
}

export interface AdminUpdateProductInput {
  name?: string;
  description?: string;
  price?: number;
  stock?: number;
  status?: ProductStatus;
  course_ids?: string[];
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  product_type: string;
  name: string;
  unit_price: number;
  qty: number;
  jumlah: number;
  weight_grams?: number;
  fulfilled_at?: string;
  created_at?: string;
}

export interface Order {
  id: string;
  student_id: string;
  status: OrderStatus;
  subtotal: number;
  discount: number;
  shipping_cost: number;
  total: number;
  promo_code_id?: string;
  shipping_address?: string;
  selected_courier?: string;
  tracking_number?: string;
  shipped_at?: string;
  gateway_ref?: string;
  payment_method?: string;
  payment_expires_at?: string;
  paid_at?: string;
  invoice_url?: string;
  estimated_delivery_days?: string;
  checked_out_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  cancellation_reason?: string;
  created_at?: string;
  updated_at?: string;
  items?: OrderItem[];
}

export interface Course {
  id: string;
  title: string;
  level?: string;
  subject?: string;
  instructor_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CourseSection {
  id: string;
  course_id: string;
  title: string;
  position?: number;
  lessons?: Lesson[];
  created_at?: string;
}

export interface Lesson {
  id: string;
  section_id: string;
  title: string;
  video_url?: string;
  duration_seconds?: number;
  position?: number;
  completed?: boolean;
  created_at?: string;
}

export interface AdminCourseDetail extends Course {
  section_count?: number;
  lesson_count?: number;
}

export interface AdminCreateCourseInput {
  title: string;
  level?: string;
  subject?: string;
  instructor_name?: string;
}

export interface AdminUpdateCourseInput {
  title?: string;
  level?: string;
  subject?: string;
  instructor_name?: string;
}

export interface AdminCreateSectionInput {
  title: string;
}

export interface AdminUpdateSectionInput {
  title: string;
}

export interface AdminCreateLessonInput {
  title: string;
  video_url?: string;
  duration?: number;
}

export interface AdminUpdateLessonInput {
  title?: string;
  video_url?: string;
  duration?: number;
}

export interface AdminReorderSectionsInput {
  section_ids: string[];
}

export interface AdminReorderLessonsInput {
  lesson_ids: string[];
}

export interface CourseSession {
  id: string;
  student_id: string;
  course_id: string;
  order_id?: string;
  status?: string;
  source?: string;
  enrolled_at?: string;
  revoked_at?: string;
  completed_lessons?: Record<string, string>;
}

export interface PromoCode {
  id: string;
  code: string;
  discount_percent?: number;
  discount_amount?: number;
  min_order_amount?: number;
  max_discount_amount?: number;
  max_uses?: number;
  used_count: number;
  expires_at?: string;
  created_at?: string;
}

export interface AdminCreatePromoCodeInput {
  code: string;
  discount_percent?: number;
  discount_amount?: number;
  max_discount_amount?: number;
  min_order_amount?: number;
  max_uses?: number;
  expires_at?: string;
}

export interface AdminUpdatePromoCodeInput {
  max_uses?: number;
  expires_at?: string;
}

export interface RevenueByTypeItem {
  total: number;
  count: number;
}

export interface AdminRevenue {
  total: number;
  by_type: Record<string, RevenueByTypeItem>;
}

export interface PromoValidation {
  code: string;
  discount: number;
  final_total: number;
}

export interface CheckoutResult {
  snap_token: string;
  gateway_ref?: string;
  payment_expires_at?: string;
}

export interface DashboardCourseSummary {
  id: string;
  title: string;
  progress: number;
  total_lessons: number;
  done_lessons: number;
  cover?: string;
}

export interface DashboardPendingOrder {
  id: string;
  product?: string;
  amount: number;
}

export interface DashboardStudySummary {
  visited_lectures: number;
  total_lectures: number;
  enrolled_courses_count: number;
  completed_courses: number;
  total_minutes: number;
}

export interface DashboardLeaderboardEntry {
  rank: number;
  name: string;
  points: number;
  is_me?: boolean;
}

export interface ExamProgressEntry {
  label: string;
  completed: number;
  in_progress: number;
}

export interface PopularLessonEntry {
  title: string;
  topics: number;
  students: number;
  duration: string;
  progress: number;
}

export interface DashboardRanking {
  position: number | null;
  points: number | null;
  leaderboard: DashboardLeaderboardEntry[];
}

export interface Dashboard {
  greeting?: string;
  enrolled_courses: DashboardCourseSummary[];
  pending_order?: DashboardPendingOrder;
  study_summary: DashboardStudySummary;
  ranking: DashboardRanking;
  exam_progress: ExamProgressEntry[];
  popular_lessons: PopularLessonEntry[];
}

export type AdminAccountRole = "super_admin" | "admin_store" | "admin_exam" | "admin_school";

export type AdminAccountStatus = "active" | "deactivated";

export interface AdminAccount {
  id: string;
  name: string;
  email?: string | null;
  role: AdminAccountRole;
  status: AdminAccountStatus;
  created_at: string;
  updated_at: string;
}

export interface AdminCreateAccountInput {
  email: string;
  name: string;
  role: AdminAccountRole;
  password: string;
}

export interface AuditLogEntry {
  id: number;
  actor_id?: string | null;
  actor_name?: string | null;
  actor_email?: string | null;
  target_type: string;
  target_id: string;
  action: string;
  metadata?: Record<string, unknown> | null;
  created_at: string;
}

export type SystemConfig = Record<string, string>;