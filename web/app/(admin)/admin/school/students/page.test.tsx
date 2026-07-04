import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import { toast } from "sonner";
import SchoolStudentsPage from "./page";
import type { AdminStudent, StudentRegistrationResult, StudentCredentials } from "@/lib/types";

const mockMutate = vi.fn();
const mockMutateAsync = vi.fn();

let studentsState = {
  data: null as { data: AdminStudent[]; next_cursor?: string } | null,
  isLoading: true,
  isFetching: false,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

let registerState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
let changeStatusState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
let reissueState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };

vi.mock("@/lib/hooks/admin-students", () => ({
  useAdminStudents: () => studentsState,
  useRegisterStudent: () => registerState,
  useChangeStudentStatus: () => changeStatusState,
  useReissueStudentCredentials: () => reissueState,
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const sampleStudents: AdminStudent[] = [
  {
    id: "st1",
    name: "Budi Santoso",
    username: "budi",
    nis: "12345",
    email: "budi@test.com",
    status: "active",
    grade: "XII IPA",
    created_at: "2026-01-15T00:00:00Z",
  },
  {
    id: "st2",
    name: "Siti Aisyah",
    username: "siti",
    nis: "67890",
    status: "deactivated",
    grade: "XI IPS",
    created_at: "2026-02-20T00:00:00Z",
  },
];

const paginatedResponse = (students: AdminStudent[]) => ({
  data: students,
  next_cursor: undefined,
});

describe("SchoolStudentsPage", () => {
  beforeEach(() => {
    studentsState = {
      data: paginatedResponse(sampleStudents),
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    registerState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    changeStatusState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    reissueState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    mockMutate.mockReset();
    mockMutateAsync.mockReset();
    (toast.success as ReturnType<typeof vi.fn>).mockReset();
    (toast.error as ReturnType<typeof vi.fn>).mockReset();
  });

  it("renders loading state when loading and no accumulated data", async () => {
    studentsState = {
      data: null,
      isLoading: true,
      isFetching: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };

    render(<SchoolStudentsPage />);

    await waitFor(() => {
      expect(screen.getByText("Memuat…")).toBeInTheDocument();
    });
  });

  it("renders error state when error and no accumulated data", async () => {
    studentsState = {
      data: null,
      isLoading: false,
      isFetching: false,
      isError: true,
      error: new Error("gagal memuat"),
      refetch: vi.fn(),
    };

    render(<SchoolStudentsPage />);

    await waitFor(() => {
      expect(screen.getByText(/gagal memuat data/i)).toBeInTheDocument();
    });
  });

  it("renders the students table with student data", async () => {
    render(<SchoolStudentsPage />);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
      expect(screen.getByText("Siti Aisyah")).toBeInTheDocument();
    });

    expect(screen.getByText(/NIS: 12345/)).toBeInTheDocument();
    expect(screen.getByText(/NIS: 67890/)).toBeInTheDocument();
    expect(screen.getByText("XII IPA")).toBeInTheDocument();
    expect(screen.getByText("XI IPS")).toBeInTheDocument();
    expect(screen.getAllByText("Aktif").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Nonaktif").length).toBeGreaterThanOrEqual(1);
  });

  it("shows stat cards with total, active, and deactivated counts", async () => {
    render(<SchoolStudentsPage />);

    await waitFor(() => {
      expect(screen.getByText("2")).toBeInTheDocument();
    });
    expect(screen.getAllByText("1").length).toBeGreaterThanOrEqual(1);
  });

  it("opens register dialog and shows credential panel on success", async () => {
    const registrationResult: StudentRegistrationResult = {
      id: "st3",
      name: "Dewi Lestari",
      username: "dewi",
      nis: "11111",
      status: "active",
      created_at: "2026-03-01T00:00:00Z",
      temp_password: "tempPass321",
    };
    mockMutateAsync.mockResolvedValueOnce(registrationResult);

    render(<SchoolStudentsPage />);

    await waitFor(() => expect(screen.getByText("Budi Santoso")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: /daftarkan siswa/i }));

    // The dialog title and submit button both say "Daftarkan Siswa"
    expect(screen.getAllByText("Daftarkan Siswa").length).toBeGreaterThanOrEqual(1);

    const nameInput = screen.getByPlaceholderText("Nama Lengkap");
    fireEvent.input(nameInput, { target: { value: "Dewi Lestari" } });

    const nisInput = screen.getByPlaceholderText("NIS");
    fireEvent.input(nisInput, { target: { value: "11111" } });

    // Click the submit button inside the dialog
    const dialog = screen.getByRole("dialog");
    const submitBtn = within(dialog).getByRole("button", { name: /daftarkan siswa/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ name: "Dewi Lestari", nis: "11111" }),
      );
      expect(toast.success).toHaveBeenCalledWith("Siswa berhasil didaftarkan.");
    });

    // Credential panel should appear
    await waitFor(() => {
      expect(screen.getByText("Kredensial Siswa")).toBeInTheDocument();
      expect(screen.getByText("dewi")).toBeInTheDocument();
      expect(screen.getByText("tempPass321")).toBeInTheDocument();
    });
  });

  it("validates required fields before register", async () => {
    render(<SchoolStudentsPage />);

    await waitFor(() => expect(screen.getByText("Budi Santoso")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: /daftarkan siswa/i }));

    // Click submit without filling required fields
    const dialog = screen.getByRole("dialog");
    const submitBtn = within(dialog).getByRole("button", { name: /daftarkan siswa/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("Semua field harus diisi");
    });

    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("renders load more button when next_cursor exists", async () => {
    studentsState = {
      data: { data: sampleStudents, next_cursor: "cursor-next" },
      isLoading: false,
      isFetching: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };

    render(<SchoolStudentsPage />);

    await waitFor(() => {
      expect(screen.getByText("Muat lebih")).toBeInTheDocument();
    });
  });

  it("surfaces an API error as error toast on register failure", async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error("gagal register"));

    render(<SchoolStudentsPage />);

    await waitFor(() => expect(screen.getByText("Budi Santoso")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: /daftarkan siswa/i }));

    const nameInput = screen.getByPlaceholderText("Nama Lengkap");
    fireEvent.input(nameInput, { target: { value: "Gagal Student" } });

    const nisInput = screen.getByPlaceholderText("NIS");
    fireEvent.input(nisInput, { target: { value: "99999" } });

    const dialog = screen.getByRole("dialog");
    const submitBtn = within(dialog).getByRole("button", { name: /daftarkan siswa/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("gagal register");
    });
  });
});
